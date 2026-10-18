// Package jobs runs restic commands for repositories as recorded jobs.
//
// Every run, scheduled or manual, goes through Manager.Enqueue: it creates a
// row in the jobs table, waits for the repository to be free (one job per
// repository at a time, in order) and for a worker slot, runs restic with a
// cancellable context, records the outcome and publishes events.
package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// Kind is what a job runs. The values match the jobs.kind column.
type Kind string

const (
	KindInit   Kind = "init"
	KindBackup Kind = "backup"
	// KindForget applies the retention policy, and prunes when the policy says so.
	KindForget Kind = "forget"
	KindPrune  Kind = "prune"
	KindCheck  Kind = "check"
	KindUnlock Kind = "unlock"
)

func (k Kind) valid() bool {
	switch k {
	case KindInit, KindBackup, KindForget, KindPrune, KindCheck, KindUnlock:
		return true
	}
	return false
}

// Trigger says why a job was started. The values match the jobs.trigger column.
type Trigger string

const (
	TriggerManual   Trigger = "manual"
	TriggerSchedule Trigger = "schedule"
)

// Status is a job's state. The values match the jobs.status column.
type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
	// StatusInterrupted: resticgui stopped while the job was queued or running.
	StatusInterrupted Status = "interrupted"
)

// Tag returns the tag put on every snapshot of a repository, so that
// retention only ever applies to snapshots resticgui made for it.
func Tag(repositoryID int64) string {
	return fmt.Sprintf("resticgui:%d", repositoryID)
}

var (
	ErrUnknownKind = errors.New("jobs: unknown job kind")
	// ErrRepositoryNotFound: the repository does not exist.
	ErrRepositoryNotFound = errors.New("jobs: repository not found")
	// ErrNotFound: the job does not exist.
	ErrNotFound = errors.New("jobs: job not found")
	// ErrFinished: the job cannot be cancelled because it already ended.
	ErrFinished = errors.New("jobs: job already finished")
	// ErrAlreadyQueued: a scheduled job was skipped because one of the same
	// kind is already queued or running for the repository.
	ErrAlreadyQueued = errors.New("jobs: a job of this kind is already queued or running")
	ErrClosed        = errors.New("jobs: manager is closed")
	// ErrNoRetention: a forget job ran for a repository without a policy.
	ErrNoRetention = errors.New("jobs: the repository has no retention policy")
)

// Restic is the part of *restic.Runner the manager uses.
type Restic interface {
	Init(ctx context.Context, repo *restic.Repo) error
	Unlock(ctx context.Context, repo *restic.Repo) error
	Backup(ctx context.Context, repo *restic.Repo, opts restic.BackupOptions, onProgress func(restic.BackupStatus)) (*restic.BackupResult, error)
	Forget(ctx context.Context, repo *restic.Repo, opts restic.ForgetOptions) ([]restic.ForgetGroup, error)
	Prune(ctx context.Context, repo *restic.Repo) error
	Check(ctx context.Context, repo *restic.Repo, opts restic.CheckOptions) (*restic.CheckSummary, error)
}

// Options configures a Manager.
type Options struct {
	// Workers is how many jobs may run at once across repositories (default 2).
	Workers int
	// ProgressInterval is the minimum time between two progress events of a
	// job (default 500ms). restic reports progress up to 60 times a second.
	ProgressInterval time.Duration
}

// Manager queues and runs jobs. It is safe for concurrent use.
type Manager struct {
	store  *store.Store
	box    *secrets.Box
	restic Restic
	events *Broker

	progressInterval time.Duration
	slots            chan struct{}

	ctx    context.Context
	stop   context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
	queues map[int64]*repoQueue
}

// repoQueue holds the jobs of one repository: at most one active, the rest
// pending in order.
type repoQueue struct {
	active  *task
	pending []*task
}

type task struct {
	job db.Job
	// cancel and cancelled are set once the task starts and guarded by Manager.mu
	cancel    context.CancelFunc
	cancelled bool
}

func New(st *store.Store, box *secrets.Box, r Restic, events *Broker, opts Options) *Manager {
	if opts.Workers <= 0 {
		opts.Workers = 2
	}
	if opts.ProgressInterval == 0 {
		opts.ProgressInterval = 500 * time.Millisecond
	}
	ctx, stop := context.WithCancel(context.Background())
	return &Manager{
		store:            st,
		box:              box,
		restic:           r,
		events:           events,
		progressInterval: opts.ProgressInterval,
		slots:            make(chan struct{}, opts.Workers),
		ctx:              ctx,
		stop:             stop,
		queues:           map[int64]*repoQueue{},
	}
}

// Recover marks the jobs left queued or running by a previous process as
// interrupted. Call it once at startup, before the first Enqueue.
func (m *Manager) Recover(ctx context.Context) (int64, error) {
	return m.store.MarkUnfinishedJobsInterrupted(ctx)
}

// Enqueue records a job and runs it once the repository and a worker are
// free. A scheduled job is skipped with ErrAlreadyQueued when the same kind
// is already queued or running for the repository.
func (m *Manager) Enqueue(ctx context.Context, repositoryID int64, kind Kind, trigger Trigger) (db.Job, error) {
	if !kind.valid() {
		return db.Job{}, fmt.Errorf("%w: %q", ErrUnknownKind, kind)
	}
	if _, err := m.store.GetRepository(ctx, repositoryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Job{}, ErrRepositoryNotFound
		}
		return db.Job{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return db.Job{}, ErrClosed
	}
	q := m.queues[repositoryID]
	if q == nil {
		q = &repoQueue{}
		m.queues[repositoryID] = q
	}
	if trigger == TriggerSchedule && q.has(kind) {
		return db.Job{}, ErrAlreadyQueued
	}
	job, err := m.store.CreateJob(ctx, db.CreateJobParams{RepositoryID: repositoryID, Kind: string(kind), Trigger: string(trigger)})
	if err != nil {
		if len(q.pending) == 0 && q.active == nil {
			delete(m.queues, repositoryID)
		}
		return db.Job{}, err
	}
	q.pending = append(q.pending, &task{job: job})
	m.publish(job, EventStatus, StatusQueued, nil)
	m.dispatch(repositoryID)
	return job, nil
}

func (q *repoQueue) has(kind Kind) bool {
	if q.active != nil && Kind(q.active.job.Kind) == kind {
		return true
	}
	for _, t := range q.pending {
		if Kind(t.job.Kind) == kind {
			return true
		}
	}
	return false
}

// dispatch starts the next pending job of a repository if none is active.
// m.mu must be held.
func (m *Manager) dispatch(repositoryID int64) {
	q := m.queues[repositoryID]
	if q == nil || q.active != nil {
		return
	}
	if len(q.pending) == 0 {
		delete(m.queues, repositoryID)
		return
	}
	t := q.pending[0]
	q.pending = q.pending[1:]
	q.active = t
	var ctx context.Context
	ctx, t.cancel = context.WithCancel(m.ctx)
	m.wg.Add(1)
	go m.run(ctx, t)
}

// Cancel stops a queued or running job. A running restic gets SIGINT, so it
// releases its repository lock before exiting.
func (m *Manager) Cancel(ctx context.Context, jobID int64) error {
	m.mu.Lock()
	for _, q := range m.queues {
		if q.active != nil && q.active.job.ID == jobID {
			q.active.cancelled = true
			q.active.cancel()
			m.mu.Unlock()
			return nil
		}
		for i, t := range q.pending {
			if t.job.ID == jobID {
				q.pending = append(q.pending[:i:i], q.pending[i+1:]...)
				m.mu.Unlock()
				return m.finish(t.job, outcome{status: StatusCancelled})
			}
		}
	}
	m.mu.Unlock()

	if _, err := m.store.GetJob(ctx, jobID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return ErrFinished
}

// Close stops accepting jobs, interrupts the running ones (restic gets
// SIGINT) and marks the queued ones interrupted, then waits for them.
func (m *Manager) Close() {
	m.mu.Lock()
	m.closed = true
	var pending []*task
	for _, q := range m.queues {
		pending = append(pending, q.pending...)
		q.pending = nil
	}
	m.mu.Unlock()
	m.stop()
	for _, t := range pending {
		_ = m.finish(t.job, outcome{status: StatusInterrupted})
	}
	m.wg.Wait()
}

func (m *Manager) run(ctx context.Context, t *task) {
	defer m.wg.Done()
	defer func() {
		t.cancel()
		m.mu.Lock()
		q := m.queues[t.job.RepositoryID]
		q.active = nil
		if !m.closed {
			m.dispatch(t.job.RepositoryID)
		} else if len(q.pending) == 0 {
			delete(m.queues, t.job.RepositoryID)
		}
		m.mu.Unlock()
	}()

	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
	case <-ctx.Done():
		_ = m.finish(t.job, m.stopped(t, nil))
		return
	}

	// the job row must be updated even while the manager shuts down
	bg := context.Background()
	if _, err := m.store.StartJob(bg, t.job.ID); err != nil {
		_ = m.finish(t.job, outcome{status: StatusFailed, err: err})
		return
	}
	m.publish(t.job, EventStatus, StatusRunning, nil)

	out := m.execute(ctx, t.job)
	if ctx.Err() != nil {
		out = m.stopped(t, out.err)
	}
	_ = m.finish(t.job, out)
}

// stopped is the outcome of a job whose context ended: cancelled by a user,
// or interrupted by Close.
func (m *Manager) stopped(t *task, err error) outcome {
	m.mu.Lock()
	cancelled := t.cancelled
	m.mu.Unlock()
	out := outcome{status: StatusInterrupted}
	if cancelled {
		out.status = StatusCancelled
	}
	var rerr *restic.Error
	if errors.As(err, &rerr) {
		out.err = err
		out.exitCode = rerr.ExitCode
		out.stderr = rerr.Stderr
	}
	return out
}

// outcome is what a job ended with, before it is written to its row.
type outcome struct {
	status     Status
	snapshotID string
	summary    any
	err        error
	exitCode   int
	stderr     []string
}

func failed(err error) outcome {
	out := outcome{status: StatusFailed, err: err}
	var rerr *restic.Error
	if errors.As(err, &rerr) {
		out.exitCode = rerr.ExitCode
		out.stderr = rerr.Stderr
	}
	return out
}

func (m *Manager) finish(job db.Job, out outcome) error {
	params := db.FinishJobParams{ID: job.ID, Status: string(out.status), SnapshotID: out.snapshotID, Summary: "{}"}
	if out.summary != nil {
		b, err := json.Marshal(out.summary)
		if err != nil {
			return err
		}
		params.Summary = string(b)
	}
	if out.err != nil && !errors.Is(out.err, context.Canceled) {
		params.Error = out.err.Error()
	}
	if out.exitCode != 0 {
		code := int64(out.exitCode)
		params.ExitCode = &code
	}
	params.LogExcerpt = strings.Join(out.stderr, "\n")
	if _, err := m.store.FinishJob(context.Background(), params); err != nil {
		return err
	}
	m.events.Publish(Event{
		Type: EventStatus, JobID: job.ID, RepositoryID: job.RepositoryID, Kind: Kind(job.Kind),
		Status: out.status, SnapshotID: out.snapshotID, Error: params.Error,
	})
	return nil
}

func (m *Manager) publish(job db.Job, typ string, status Status, p *Progress) {
	m.events.Publish(Event{Type: typ, JobID: job.ID, RepositoryID: job.RepositoryID, Kind: Kind(job.Kind), Status: status, Progress: p})
}

// execute runs the restic command of a job.
func (m *Manager) execute(ctx context.Context, job db.Job) outcome {
	repo, err := m.store.GetRepository(ctx, job.RepositoryID)
	if err != nil {
		return failed(err)
	}
	target, err := m.resticRepo(repo)
	if err != nil {
		return failed(err)
	}

	switch Kind(job.Kind) {
	case KindInit:
		if err := m.restic.Init(ctx, target); err != nil {
			return failed(err)
		}
		if err := m.store.MarkRepositoryInitialized(context.Background(), repo.ID); err != nil {
			return failed(err)
		}
	case KindBackup:
		return m.backup(ctx, job, repo, target)
	case KindForget:
		return m.forget(ctx, repo, target)
	case KindPrune:
		if err := m.restic.Prune(ctx, target); err != nil {
			return failed(err)
		}
	case KindCheck:
		summary, err := m.restic.Check(ctx, target, restic.CheckOptions{})
		if err != nil {
			out := failed(err)
			out.summary = summary
			return out
		}
		return outcome{status: StatusSucceeded, summary: summary}
	case KindUnlock:
		if err := m.restic.Unlock(ctx, target); err != nil {
			return failed(err)
		}
	default:
		return failed(fmt.Errorf("%w: %q", ErrUnknownKind, job.Kind))
	}
	return outcome{status: StatusSucceeded}
}

func (m *Manager) resticRepo(repo db.Repository) (*restic.Repo, error) {
	sec, err := store.OpenSecrets(m.box, repo.Secrets)
	if err != nil {
		return nil, fmt.Errorf("repository %q: decrypt secrets: %w", repo.Name, err)
	}
	opts, err := store.DecodeOptions(repo.Options)
	if err != nil {
		return nil, fmt.Errorf("repository %q: decode options: %w", repo.Name, err)
	}
	env := map[string]string{}
	for k, v := range opts.Env {
		env[k] = v
	}
	for k, v := range sec.Env {
		env[k] = v
	}
	return &restic.Repo{Location: repo.Location, Password: sec.Password, Env: env, CACert: opts.CACert, Options: opts.Extended}, nil
}

func (m *Manager) backup(ctx context.Context, job db.Job, repo db.Repository, target *restic.Repo) outcome {
	paths, err := m.store.ListPaths(ctx, repo.ID)
	if err != nil {
		return failed(err)
	}
	opts := restic.BackupOptions{
		Tags:              []string{Tag(repo.ID)},
		Host:              repo.Host,
		ExcludeLargerThan: repo.ExcludeLargerThan,
		Compression:       repo.Compression,
	}
	for _, p := range paths {
		if p.Kind == "exclude" {
			opts.Excludes = append(opts.Excludes, p.Path)
		} else {
			opts.Paths = append(opts.Paths, p.Path)
		}
	}

	var last time.Time
	result, err := m.restic.Backup(ctx, target, opts, func(s restic.BackupStatus) {
		if now := time.Now(); now.Sub(last) >= m.progressInterval {
			last = now
			m.publish(job, EventProgress, StatusRunning, &Progress{
				PercentDone: s.PercentDone, FilesDone: s.FilesDone, TotalFiles: s.TotalFiles,
				BytesDone: s.BytesDone, TotalBytes: s.TotalBytes,
				SecondsRemaining: s.SecondsRemaining, CurrentFiles: s.CurrentFiles,
			})
		}
	})
	var out outcome
	switch {
	case err == nil:
		out = outcome{status: StatusSucceeded}
	case errors.Is(err, restic.ErrIncomplete):
		// the snapshot exists without the unreadable files: the job succeeded,
		// and its error says what was skipped
		out = failed(err)
		out.status = StatusSucceeded
		if result != nil && len(result.Errors) > 0 {
			out.err = fmt.Errorf("%d files could not be read, first: %s: %s", len(result.Errors), result.Errors[0].Item, result.Errors[0].Message)
		}
	default:
		return failed(err)
	}
	if result != nil && result.Summary != nil {
		out.summary = result.Summary
		out.snapshotID = result.Summary.SnapshotID
	}
	return out
}

// ForgetSummary is the summary recorded for a forget job.
type ForgetSummary struct {
	Kept    int  `json:"kept"`
	Removed int  `json:"removed"`
	Pruned  bool `json:"pruned"`
}

func (m *Manager) forget(ctx context.Context, repo db.Repository, target *restic.Repo) outcome {
	ret, err := m.store.GetRetention(ctx, repo.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return failed(ErrNoRetention)
	}
	if err != nil {
		return failed(err)
	}
	opts := restic.ForgetOptions{
		Policy: restic.Policy{
			KeepLast: int(ret.KeepLast), KeepHourly: int(ret.KeepHourly), KeepDaily: int(ret.KeepDaily),
			KeepWeekly: int(ret.KeepWeekly), KeepMonthly: int(ret.KeepMonthly), KeepYearly: int(ret.KeepYearly),
			KeepWithin: ret.KeepWithin,
		},
		// only snapshots resticgui made for this repository, grouped so that
		// editing the backup paths does not start a new retention group
		Filter:  restic.SnapshotFilter{Host: repo.Host, Tags: []string{Tag(repo.ID)}},
		GroupBy: "host,tags",
		Prune:   ret.Prune,
	}
	if opts.Policy.IsEmpty() {
		return failed(ErrNoRetention)
	}
	groups, err := m.restic.Forget(ctx, target, opts)
	if err != nil {
		return failed(err)
	}
	summary := ForgetSummary{Pruned: ret.Prune}
	for _, g := range groups {
		summary.Kept += len(g.Keep)
		summary.Removed += len(g.Remove)
	}
	return outcome{status: StatusSucceeded, summary: summary}
}
