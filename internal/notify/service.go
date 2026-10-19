package notify

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/scheduler"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// What a notifier can be told about.
const (
	EventFailure = "failure"
	EventSuccess = "success"
	EventMissed  = "missed"
)

// Grace is how late a scheduled backup may be before it counts as missed.
const Grace = time.Hour

// Service watches the job events, sends notifications and pings
// Healthchecks. It also checks every few minutes for scheduled backups
// that did not happen.
type Service struct {
	Store  *store.Store
	Box    *secrets.Box
	Events *jobs.Broker
	// Busy reports whether a job is queued or running for a repository.
	Busy func(repositoryID int64) bool
	// Send delivers notifications; default Shoutrrr.
	Send Sender
	// HTTP pings Healthchecks; default a client with a 10s timeout.
	HTTP *http.Client
	// Now is the clock; default time.Now.
	Now func() time.Time
	// CheckEvery is how often missed backups are looked for; default 5 minutes.
	CheckEvery time.Duration

	mu sync.Mutex
	// alerted holds, per plan, the due time of the last missed backup
	// already reported, so that it is reported once.
	alerted map[int64]time.Time
	wg      sync.WaitGroup
}

func (s *Service) init() {
	if s.Send == nil {
		s.Send = Shoutrrr
	}
	if s.HTTP == nil {
		s.HTTP = &http.Client{Timeout: 10 * time.Second}
	}
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.CheckEvery == 0 {
		s.CheckEvery = 5 * time.Minute
	}
	if s.Busy == nil {
		s.Busy = func(int64) bool { return false }
	}
	s.mu.Lock()
	if s.alerted == nil {
		s.alerted = map[int64]time.Time{}
	}
	s.mu.Unlock()
}

// Run handles events until ctx ends, then waits for the messages being sent.
func (s *Service) Run(ctx context.Context) {
	s.init()
	defer s.wg.Wait()
	tick := time.NewTicker(s.CheckEvery)
	defer tick.Stop()
	for {
		events, unsubscribe := s.Events.Subscribe(256)
		if !s.loop(ctx, events, tick.C) {
			unsubscribe()
			return
		}
		// the broker dropped a subscriber that fell behind: subscribe again
		unsubscribe()
	}
}

func (s *Service) loop(ctx context.Context, events <-chan jobs.Event, tick <-chan time.Time) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case <-tick:
			s.CheckMissed(ctx)
		case e, ok := <-events:
			if !ok {
				return true
			}
			if e.Type == jobs.EventStatus {
				s.wg.Add(1)
				// sending can take seconds; jobs never wait for it
				go func() {
					defer s.wg.Done()
					s.Handle(context.WithoutCancel(ctx), e)
				}()
			}
		}
	}
}

var alerted = map[jobs.Kind]bool{jobs.KindBackup: true, jobs.KindForget: true, jobs.KindPrune: true, jobs.KindCheck: true}

// Handle reacts to one job status event.
func (s *Service) Handle(ctx context.Context, e jobs.Event) {
	s.init()
	if !alerted[e.Kind] {
		return
	}
	repo, err := s.Store.GetRepository(ctx, e.RepositoryID)
	if err != nil {
		return
	}
	if e.Kind == jobs.KindBackup {
		s.ping(ctx, repo, e)
	}
	var event string
	switch e.Status {
	case jobs.StatusSucceeded:
		event = EventSuccess
	case jobs.StatusFailed, jobs.StatusInterrupted:
		event = EventFailure
	default:
		return
	}
	job, err := s.Store.GetJob(ctx, e.JobID)
	if err != nil {
		return
	}
	title, body := describe(s.subject(ctx, repo, job.PlanID), job)
	s.notify(ctx, event, title, body)
}

// subject names what a job ran on: the repository, and the plan when the
// repository has several.
func (s *Service) subject(ctx context.Context, repo db.Repository, planID *int64) string {
	if planID == nil {
		return repo.Name
	}
	plans, err := s.Store.ListPlans(ctx, repo.ID)
	if err != nil || len(plans) < 2 {
		return repo.Name
	}
	for _, p := range plans {
		if p.ID == *planID {
			return fmt.Sprintf("%s (%s)", repo.Name, p.Name)
		}
	}
	return repo.Name
}

// notify sends to every enabled notifier that wants event.
func (s *Service) notify(ctx context.Context, event, title, body string) {
	notifiers, err := s.Store.ListNotifiers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("notifications: list the notifiers")
		return
	}
	for _, n := range notifiers {
		if !n.Enabled || !wants(n, event) {
			continue
		}
		if err := s.SendTo(n, title, body); err != nil {
			log.Warn().Err(err).Str("notifier", n.Name).Msg("notification not sent")
		}
	}
}

func wants(n db.Notifier, event string) bool {
	switch event {
	case EventFailure:
		return n.OnFailure
	case EventSuccess:
		return n.OnSuccess
	case EventMissed:
		return n.OnMissed
	}
	return false
}

// SendTo sends one message through a stored notifier.
func (s *Service) SendTo(n db.Notifier, title, body string) error {
	s.init()
	rawURL, err := s.Box.Open(n.Url)
	if err != nil {
		return fmt.Errorf("decrypt the URL: %w", err)
	}
	return s.Send(string(rawURL), title, body)
}

// ping tells the repository's Healthchecks check that a backup started,
// succeeded or failed.
func (s *Service) ping(ctx context.Context, repo db.Repository, e jobs.Event) {
	hc, err := s.Store.GetHealthcheck(ctx, repo.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("healthchecks: load the check")
		return
	}
	var suffix, body string
	switch e.Status {
	case jobs.StatusRunning:
		suffix = "/start"
	case jobs.StatusSucceeded:
		body = "Backup succeeded, snapshot " + e.SnapshotID
		if e.Error != "" {
			body += ": " + e.Error
		}
	case jobs.StatusFailed, jobs.StatusInterrupted:
		suffix, body = "/fail", "Backup "+string(e.Status)+": "+e.Error
	default:
		return
	}
	raw, err := s.Box.Open(hc.Url)
	if err != nil {
		log.Error().Err(err).Msg("healthchecks: decrypt the URL")
		return
	}
	if err := Ping(ctx, s.HTTP, strings.TrimRight(string(raw), "/")+suffix, body); err != nil {
		log.Warn().Err(Redact(err, string(raw))).Str("repository", repo.Name).Msg("healthchecks ping failed")
	}
}

// Ping sends a Healthchecks ping, with body as the run's log.
func Ping(ctx context.Context, client *http.Client, pingURL, body string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pingURL, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ResticGUI")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("answered %s", res.Status)
	}
	return nil
}

var kindNames = map[jobs.Kind]string{jobs.KindBackup: "Backup", jobs.KindForget: "Retention", jobs.KindPrune: "Prune", jobs.KindCheck: "Check"}

// describe returns a notification's title and body for a finished job of
// what subject names.
func describe(subject string, job db.Job) (string, string) {
	name := kindNames[jobs.Kind(job.Kind)]
	switch jobs.Status(job.Status) {
	case jobs.StatusSucceeded:
		body := fmt.Sprintf("%s of %s succeeded.", name, subject)
		var sum restic.BackupSummary
		if job.Kind == string(jobs.KindBackup) && json.Unmarshal([]byte(job.Summary), &sum) == nil && sum.SnapshotID != "" {
			body = fmt.Sprintf("Backup of %s succeeded: snapshot %.8s, %d new and %d changed files, %s added in %s.",
				subject, sum.SnapshotID, sum.FilesNew, sum.FilesChanged, Bytes(sum.DataAdded), (time.Duration(sum.TotalDuration * float64(time.Second))).Round(time.Second))
		}
		if job.Error != "" {
			body += "\n" + job.Error
		}
		return fmt.Sprintf("✅ %s of %s succeeded", name, subject), body
	case jobs.StatusInterrupted:
		return fmt.Sprintf("⚠️ %s of %s was interrupted", name, subject),
			fmt.Sprintf("%s of %s stopped because ResticGUI stopped. It runs again at its next scheduled time.", name, subject)
	default:
		body := fmt.Sprintf("%s of %s failed", name, subject)
		if job.Error != "" {
			body += ": " + job.Error
		}
		return fmt.Sprintf("❌ %s of %s failed", name, subject), body + "."
	}
}

// Bytes formats a size the way the web UI does.
func Bytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// CheckMissed reports, once each, the scheduled backups of plans that are
// more than Grace late with no backup of the plan run since they were due.
func (s *Service) CheckMissed(ctx context.Context) {
	s.init()
	schedules, err := s.Store.ListEnabledPlanSchedules(ctx)
	if err != nil {
		return
	}
	last, err := s.Store.LastPlanBackups(ctx)
	if err != nil {
		return
	}
	// the end of the last backup that ran (whatever its outcome), per plan
	ran := map[int64]time.Time{}
	for _, j := range last {
		if j.FinishedAt != nil {
			ran[*j.PlanID] = *j.FinishedAt
		}
	}
	now := s.Now()
	for _, sc := range schedules {
		if sc.Job != string(jobs.KindBackup) || s.Busy(sc.RepositoryID) {
			continue
		}
		spec, err := scheduler.Parse(sc.Cron)
		if err != nil {
			continue
		}
		plan, err := s.Store.GetPlan(ctx, sc.PlanID)
		if err != nil {
			continue
		}
		since := plan.CreatedAt
		if t, ok := ran[plan.ID]; ok {
			since = t
		}
		due := spec.Next(since)
		if !Missed(due, now) {
			continue
		}
		s.mu.Lock()
		already := s.alerted[plan.ID].Equal(due)
		s.alerted[plan.ID] = due
		s.mu.Unlock()
		if already {
			continue
		}
		repo, err := s.Store.GetRepository(ctx, sc.RepositoryID)
		if err != nil {
			continue
		}
		subject := s.subject(ctx, repo, &plan.ID)
		title := fmt.Sprintf("⏰ Backup of %s did not run", subject)
		body := fmt.Sprintf("The backup of %s was due at %s and has not run. Is ResticGUI able to reach the repository, and was it running then?",
			subject, due.Local().Format("Mon 2 Jan 15:04"))
		s.notify(ctx, EventMissed, title, body)
	}
}

// Missed reports whether a backup due at due is more than Grace late at now.
func Missed(due, now time.Time) bool {
	return !due.IsZero() && now.Sub(due) > Grace
}
