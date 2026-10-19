// Package scheduler enqueues the jobs of the repositories' and plans'
// schedules when their cron expression comes due.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// Parse parses a schedule: five cron fields ("30 2 * * *") or a descriptor
// such as "@daily" or "@every 6h". Use it to validate user input.
func Parse(spec string) (cron.Schedule, error) {
	s, err := parser.Parse(spec)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule %q: %w", spec, err)
	}
	return s, nil
}

// Clock is the time source of the scheduler, replaced in tests.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// Enqueuer starts jobs; *jobs.Manager implements it.
type Enqueuer interface {
	Enqueue(ctx context.Context, repositoryID, planID int64, kind jobs.Kind, trigger jobs.Trigger) (db.Job, error)
}

// Source lists the enabled schedules; *store.Store implements it.
type Source interface {
	EnabledSchedules(ctx context.Context) ([]store.Schedule, error)
}

// Scheduler runs the enabled schedules. Runs missed while resticgui was
// stopped are not caught up: each schedule next fires at its next due time.
type Scheduler struct {
	source  Source
	enqueue Enqueuer
	clock   Clock
	reload  chan struct{}
}

func New(source Source, enqueue Enqueuer, clock Clock) *Scheduler {
	if clock == nil {
		clock = realClock{}
	}
	return &Scheduler{source: source, enqueue: enqueue, clock: clock, reload: make(chan struct{}, 1)}
}

// Reload makes Run re-read the schedules, after one was added, changed or
// removed. It never blocks.
func (s *Scheduler) Reload() {
	select {
	case s.reload <- struct{}{}:
	default:
	}
}

type entry struct {
	schedule store.Schedule
	spec     cron.Schedule
	next     time.Time
}

// Run fires schedules until ctx ends.
func (s *Scheduler) Run(ctx context.Context) error {
	entries := s.load(ctx)
	for {
		var wake <-chan time.Time
		if next, ok := earliest(entries); ok {
			wake = s.clock.After(next.Sub(s.clock.Now()))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.reload:
			entries = s.load(ctx)
		case <-wake:
			now := s.clock.Now()
			for i := range entries {
				if e := &entries[i]; !e.next.IsZero() && !e.next.After(now) {
					s.fire(ctx, e.schedule)
					e.next = e.spec.Next(now)
				}
			}
		}
	}
}

func (s *Scheduler) load(ctx context.Context) []entry {
	schedules, err := s.source.EnabledSchedules(ctx)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: cannot list schedules")
		return nil
	}
	now := s.clock.Now()
	entries := make([]entry, 0, len(schedules))
	for _, sc := range schedules {
		spec, err := Parse(sc.Cron)
		if err != nil {
			log.Error().Err(err).Int64("repository", sc.RepositoryID).Int64("plan", sc.PlanID).Str("job", sc.Job).Msg("scheduler: skipping schedule")
			continue
		}
		entries = append(entries, entry{schedule: sc, spec: spec, next: spec.Next(now)})
	}
	return entries
}

func earliest(entries []entry) (time.Time, bool) {
	var t time.Time
	for _, e := range entries {
		if !e.next.IsZero() && (t.IsZero() || e.next.Before(t)) {
			t = e.next
		}
	}
	return t, !t.IsZero()
}

func (s *Scheduler) fire(ctx context.Context, sc store.Schedule) {
	job, err := s.enqueue.Enqueue(ctx, sc.RepositoryID, sc.PlanID, jobs.Kind(sc.Job), jobs.TriggerSchedule)
	l := log.With().Int64("repository", sc.RepositoryID).Int64("plan", sc.PlanID).Str("job", sc.Job).Logger()
	switch {
	case errors.Is(err, jobs.ErrAlreadyQueued):
		l.Warn().Msg("scheduler: previous run still queued or running, skipping")
	case err != nil:
		l.Error().Err(err).Msg("scheduler: cannot enqueue")
	default:
		l.Debug().Int64("job_id", job.ID).Msg("scheduler: enqueued")
	}
}
