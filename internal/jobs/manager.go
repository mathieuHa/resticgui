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
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	zlog "github.com/rs/zerolog/log"

	"github.com/mathieuHa/resticgui/internal/hooks"
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
	// KindRestore restores a snapshot, as described by its RestoreParams.
	KindRestore Kind = "restore"
	// KindRepairIndex builds the index again from the pack files.
	KindRepairIndex Kind = "repair_index"
	// KindRepairSnapshots rewrites the snapshots that refer to missing data
	// without it, and forgets the damaged ones.
	KindRepairSnapshots Kind = "repair_snapshots"
	// KindMigrate upgrades the repository to format version 2, when it is 1.
	KindMigrate Kind = "migrate"
	// KindChangePassword changes the repository's password, as given to
	// EnqueuePasswordChange.
	KindChangePassword Kind = "change_password"
)

func (k Kind) valid() bool {
	switch k {
	case KindInit, KindBackup, KindForget, KindPrune, KindCheck, KindUnlock, KindRestore,
		KindRepairIndex, KindRepairSnapshots, KindMigrate, KindChangePassword:
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

// Tag returns the tag of the snapshots resticgui made for a repository
// before it had plans, which its first plan kept. The snapshots of the
// plans made later have tags starting with Tag and a colon.
func Tag(repositoryID int64) string {
	return fmt.Sprintf("resticgui:%d", repositoryID)
}

// OfRepository reports whether a snapshot with these tags was made by
// resticgui for the repository, by any of its plans, also deleted ones.
func OfRepository(repositoryID int64, tags []string) bool {
	tag := Tag(repositoryID)
	for _, t := range tags {
		if t == tag || strings.HasPrefix(t, tag+":") {
			return true
		}
	}
	return false
}

// byPlan reports whether jobs of kind belong to a plan.
func byPlan(kind Kind) bool {
	return kind == KindBackup || kind == KindForget
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
	// ErrNoRetention: a forget job ran for a plan without a policy.
	ErrNoRetention = errors.New("jobs: the plan has no retention policy")
	// ErrPlanNotFound: the plan does not exist, or is not the repository's.
	ErrPlanNotFound = errors.New("jobs: plan not found")
	// ErrPlanRequired: a backup or forget of a repository with several
	// plans, or none, does not say which plan.
	ErrPlanRequired = errors.New("jobs: say which plan to run")
)

// Restic is the part of *restic.Runner the manager uses.
type Restic interface {
	Init(ctx context.Context, repo *restic.Repo) error
	Unlock(ctx context.Context, repo *restic.Repo) error
	Backup(ctx context.Context, repo *restic.Repo, opts restic.BackupOptions, onProgress func(restic.BackupStatus)) (*restic.BackupResult, error)
	Forget(ctx context.Context, repo *restic.Repo, opts restic.ForgetOptions) ([]restic.ForgetGroup, error)
	Prune(ctx context.Context, repo *restic.Repo) error
	Check(ctx context.Context, repo *restic.Repo, opts restic.CheckOptions) (*restic.CheckSummary, error)
	Restore(ctx context.Context, repo *restic.Repo, opts restic.RestoreOptions, onProgress func(restic.RestoreStatus)) (*restic.RestoreSummary, error)
	Stats(ctx context.Context, repo *restic.Repo, mode, snapshotID string) (*restic.Stats, error)
	RepairIndex(ctx context.Context, repo *restic.Repo) error
	RepairSnapshots(ctx context.Context, repo *restic.Repo) (int, error)
	Migrations(ctx context.Context, repo *restic.Repo) ([]string, error)
	Migrate(ctx context.Context, repo *restic.Repo, name string) error
	Keys(ctx context.Context, repo *restic.Repo) ([]restic.Key, error)
	AddKey(ctx context.Context, repo *restic.Repo, password string) error
	RemoveKey(ctx context.Context, repo *restic.Repo, id string) error
}

// RestoreParams says what a restore job restores, and where.
type RestoreParams struct {
	SnapshotID string   `json:"snapshot_id"`
	Target     string   `json:"target"`
	Include    []string `json:"include,omitempty"`
}

// CheckParams says how much data a check job reads.
type CheckParams struct {
	// ReadDataSubset is restic's --read-data-subset: "10%", "1/5" or a size
	// such as "500M". Empty reads no data, only the structure.
	ReadDataSubset string `json:"read_data_subset,omitempty"`
}

var subsetPattern = regexp.MustCompile(`^([1-9][0-9]*/[1-9][0-9]*|[0-9]+(\.[0-9]+)?%|[1-9][0-9]*[KMGT]?)$`)

// ValidSubset checks a --read-data-subset value; empty is valid.
func ValidSubset(subset string) error {
	if subset == "" || subsetPattern.MatchString(subset) {
		return nil
	}
	return fmt.Errorf("read_data_subset %q is not a share such as 10%%, a part such as 1/5, or a size such as 500M", subset)
}

// Options configures a Manager.
type Options struct {
	// Workers is how many jobs may run at once across repositories (default 2).
	Workers int
	// ProgressInterval is the minimum time between two progress events of a
	// job (default 500ms). restic reports progress up to 60 times a second.
	ProgressInterval time.Duration
	// Hooks runs the hooks set on repositories around backups.
	Hooks hooks.Runner
}

// Manager queues and runs jobs. It is safe for concurrent use.
type Manager struct {
	store  *store.Store
	box    *secrets.Box
	restic Restic
	events *Broker
	hooks  hooks.Runner

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
	// password is the new password of a change_password job, kept in memory
	// only: a job interrupted by a restart is not run again
	password string
	// cancel and cancelled are set once the task starts and guarded by Manager.mu
	cancel    context.CancelFunc
	cancelled bool
	// finished is set, under Manager.mu, once the job's outcome is being
	// written, so that a cancel arriving then answers ErrFinished
	finished bool
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
		hooks:            opts.Hooks,
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
// is already queued or running for the repository, or for the plan.
//
// A backup or forget runs a plan: planID, or the repository's only plan
// when planID is 0.
func (m *Manager) Enqueue(ctx context.Context, repositoryID, planID int64, kind Kind, trigger Trigger) (db.Job, error) {
	switch kind {
	case KindRestore:
		return db.Job{}, fmt.Errorf("%w: a restore needs EnqueueRestore", ErrUnknownKind)
	case KindChangePassword:
		return db.Job{}, fmt.Errorf("%w: a password change needs EnqueuePasswordChange", ErrUnknownKind)
	case KindCheck:
		if trigger == TriggerSchedule {
			return m.scheduledCheck(ctx, repositoryID)
		}
	}
	var plan *int64
	if byPlan(kind) {
		id, err := m.plan(ctx, repositoryID, planID)
		if err != nil {
			return db.Job{}, err
		}
		plan = &id
	}
	return m.enqueue(ctx, repositoryID, plan, kind, trigger, "{}", "")
}

// scheduledCheck queues the check of a repository's schedule, reading as
// much data as the schedule says.
func (m *Manager) scheduledCheck(ctx context.Context, repositoryID int64) (db.Job, error) {
	var p CheckParams
	schedules, err := m.store.ListSchedules(ctx, repositoryID)
	if err != nil {
		return db.Job{}, err
	}
	for _, s := range schedules {
		if Kind(s.Job) == KindCheck {
			p.ReadDataSubset = s.ReadDataSubset
		}
	}
	return m.enqueueCheck(ctx, repositoryID, TriggerSchedule, p)
}

// EnqueueCheck queues a manual check that reads some of the data.
func (m *Manager) EnqueueCheck(ctx context.Context, repositoryID int64, p CheckParams) (db.Job, error) {
	return m.enqueueCheck(ctx, repositoryID, TriggerManual, p)
}

func (m *Manager) enqueueCheck(ctx context.Context, repositoryID int64, trigger Trigger, p CheckParams) (db.Job, error) {
	if err := ValidSubset(p.ReadDataSubset); err != nil {
		return db.Job{}, err
	}
	params, err := json.Marshal(p)
	if err != nil {
		return db.Job{}, err
	}
	return m.enqueue(ctx, repositoryID, nil, KindCheck, trigger, string(params), "")
}

// EnqueuePasswordChange queues the change of a repository's password. The
// job adds a key for the new password, saves it in place of the old one,
// then removes the old key.
func (m *Manager) EnqueuePasswordChange(ctx context.Context, repositoryID int64, password string) (db.Job, error) {
	if password == "" {
		return db.Job{}, errors.New("jobs: the new password is empty")
	}
	return m.enqueue(ctx, repositoryID, nil, KindChangePassword, TriggerManual, "{}", password)
}

// plan checks that planID is a plan of the repository, or finds its only
// plan when planID is 0.
func (m *Manager) plan(ctx context.Context, repositoryID, planID int64) (int64, error) {
	if planID != 0 {
		p, err := m.store.GetPlan(ctx, planID)
		if errors.Is(err, sql.ErrNoRows) || err == nil && p.RepositoryID != repositoryID {
			return 0, ErrPlanNotFound
		}
		return planID, err
	}
	if _, err := m.store.GetRepository(ctx, repositoryID); errors.Is(err, sql.ErrNoRows) {
		return 0, ErrRepositoryNotFound
	}
	plans, err := m.store.ListPlans(ctx, repositoryID)
	if err != nil {
		return 0, err
	}
	if len(plans) != 1 {
		return 0, fmt.Errorf("%w: the repository has %d plans", ErrPlanRequired, len(plans))
	}
	return plans[0].ID, nil
}

// EnqueueRestore queues a manual restore of a snapshot.
func (m *Manager) EnqueueRestore(ctx context.Context, repositoryID int64, p RestoreParams) (db.Job, error) {
	if p.SnapshotID == "" || p.Target == "" {
		return db.Job{}, errors.New("jobs: a restore needs a snapshot and a target")
	}
	params, err := json.Marshal(p)
	if err != nil {
		return db.Job{}, err
	}
	return m.enqueue(ctx, repositoryID, nil, KindRestore, TriggerManual, string(params), "")
}

func (m *Manager) enqueue(ctx context.Context, repositoryID int64, planID *int64, kind Kind, trigger Trigger, params, password string) (db.Job, error) {
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
	if trigger == TriggerSchedule && q.has(kind, planID) {
		return db.Job{}, ErrAlreadyQueued
	}
	job, err := m.store.CreateJob(ctx, db.CreateJobParams{RepositoryID: repositoryID, PlanID: planID, Kind: string(kind), Trigger: string(trigger), Params: params})
	if err != nil {
		if len(q.pending) == 0 && q.active == nil {
			delete(m.queues, repositoryID)
		}
		return db.Job{}, err
	}
	q.pending = append(q.pending, &task{job: job, password: password})
	m.publish(job, EventStatus, StatusQueued, nil)
	m.dispatch(repositoryID)
	return job, nil
}

// Busy reports whether a job is queued or running for the repository.
func (m *Manager) Busy(repositoryID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	q := m.queues[repositoryID]
	return q != nil && (q.active != nil || len(q.pending) > 0)
}

func (q *repoQueue) has(kind Kind, planID *int64) bool {
	same := func(j db.Job) bool {
		return Kind(j.Kind) == kind && (planID == nil || j.PlanID != nil && *j.PlanID == *planID)
	}
	if q.active != nil && same(q.active.job) {
		return true
	}
	for _, t := range q.pending {
		if same(t.job) {
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
		if q.active != nil && q.active.job.ID == jobID && !q.active.finished {
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

	finish := func(out outcome) {
		m.mu.Lock()
		t.finished = true
		m.mu.Unlock()
		_ = m.finish(t.job, out)
	}

	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
	case <-ctx.Done():
		finish(m.stopped(t, nil))
		return
	}

	// the job row must be updated even while the manager shuts down
	bg := context.Background()
	if _, err := m.store.StartJob(bg, t.job.ID); err != nil {
		finish(outcome{status: StatusFailed, err: err})
		return
	}
	m.publish(t.job, EventStatus, StatusRunning, nil)

	out := m.execute(ctx, t.job, t.password)
	// a password change runs to its end, so that the saved password is
	// always one that opens the repository
	if ctx.Err() != nil && Kind(t.job.Kind) != KindChangePassword {
		log := out.stderr
		out = m.stopped(t, out.err)
		if len(out.stderr) == 0 {
			// keep what the hooks printed
			out.stderr = log
		}
	}
	finish(out)
	if out.status == StatusSucceeded && changesSnapshots(Kind(t.job.Kind)) {
		m.refreshStats(t.job.RepositoryID)
	}
}

// changesSnapshots reports whether a job of kind changes the snapshot count
// or the size of the repository.
func changesSnapshots(kind Kind) bool {
	return kind == KindBackup || kind == KindForget || kind == KindPrune || kind == KindInit || kind == KindRepairSnapshots
}

// refreshStats records the snapshot count and size of a repository for the
// metrics endpoint. It runs while the job still holds the repository, so
// that it never waits for a lock, and a failure only leaves the old values.
func (m *Manager) refreshStats(repositoryID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	repo, err := m.store.GetRepository(ctx, repositoryID)
	if err != nil {
		return
	}
	target, err := ResticRepo(m.box, repo)
	if err != nil {
		return
	}
	stats, err := m.restic.Stats(ctx, target, "raw-data", "")
	if err != nil {
		zlog.Warn().Err(err).Str("repository", repo.Name).Msg("could not refresh the repository size")
		return
	}
	_ = m.store.UpsertRepositoryStats(ctx, db.UpsertRepositoryStatsParams{
		RepositoryID: repositoryID, Snapshots: int64(stats.SnapshotsCount), TotalSize: int64(stats.TotalSize),
	})
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
	if out.status == StatusSucceeded && Kind(job.Kind) != KindUnlock {
		// the command opened the repository, so it exists, also when it was
		// created outside resticgui and never had an init job
		if err := m.store.MarkRepositoryInitialized(context.Background(), job.RepositoryID); err != nil {
			return err
		}
	}
	if _, err := m.store.FinishJob(context.Background(), params); err != nil {
		return err
	}
	m.events.Publish(Event{
		Type: EventStatus, JobID: job.ID, RepositoryID: job.RepositoryID, PlanID: job.PlanID, Kind: Kind(job.Kind),
		Status: out.status, SnapshotID: out.snapshotID, Error: params.Error,
	})
	return nil
}

func (m *Manager) publish(job db.Job, typ string, status Status, p *Progress) {
	m.events.Publish(Event{Type: typ, JobID: job.ID, RepositoryID: job.RepositoryID, PlanID: job.PlanID, Kind: Kind(job.Kind), Status: status, Progress: p})
}

// execute runs the restic command of a job; password is the new password of
// a change_password job.
func (m *Manager) execute(ctx context.Context, job db.Job, password string) outcome {
	repo, err := m.store.GetRepository(ctx, job.RepositoryID)
	if err != nil {
		return failed(err)
	}
	target, err := ResticRepo(m.box, repo)
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
	case KindBackup, KindForget:
		if job.PlanID == nil {
			return failed(ErrPlanRequired)
		}
		plan, err := m.store.GetPlan(ctx, *job.PlanID)
		if errors.Is(err, sql.ErrNoRows) {
			return failed(fmt.Errorf("%w: it was deleted", ErrPlanNotFound))
		}
		if err != nil {
			return failed(err)
		}
		if Kind(job.Kind) == KindForget {
			return m.forget(ctx, repo, plan, target)
		}
		return m.withHooks(ctx, job, repo, plan, func() outcome { return m.backup(ctx, job, repo, plan, target) })
	case KindPrune:
		if err := m.restic.Prune(ctx, target); err != nil {
			return failed(err)
		}
	case KindCheck:
		var p CheckParams
		if err := json.Unmarshal([]byte(job.Params), &p); err != nil {
			return failed(fmt.Errorf("check parameters: %w", err))
		}
		summary, err := m.restic.Check(ctx, target, restic.CheckOptions{ReadDataSubset: p.ReadDataSubset})
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
	case KindRestore:
		return m.restore(ctx, job, target)
	case KindRepairIndex:
		if err := m.restic.RepairIndex(ctx, target); err != nil {
			return failed(err)
		}
	case KindRepairSnapshots:
		n, err := m.restic.RepairSnapshots(ctx, target)
		if err != nil {
			return failed(err)
		}
		return outcome{status: StatusSucceeded, summary: RepairSummary{Modified: n}}
	case KindMigrate:
		return m.migrate(ctx, target)
	case KindChangePassword:
		// not cancellable: see run
		return m.changePassword(context.WithoutCancel(ctx), repo, target, password)
	default:
		return failed(fmt.Errorf("%w: %q", ErrUnknownKind, job.Kind))
	}
	return outcome{status: StatusSucceeded}
}

// ResticRepo builds what restic needs to open a stored repository: its
// location, decrypted password and variables, and options.
func ResticRepo(box *secrets.Box, repo db.Repository) (*restic.Repo, error) {
	sec, err := store.OpenSecrets(box, repo.Secrets)
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

// BackupOptions is how a plan of repo runs restic backup.
func BackupOptions(repo db.Repository, plan db.Plan, paths []db.PlanPath) (restic.BackupOptions, error) {
	opts, err := store.DecodePlanOptions(plan.Options)
	if err != nil {
		return restic.BackupOptions{}, fmt.Errorf("plan %q: decode options: %w", plan.Name, err)
	}
	tags, err := store.DecodeTags(plan.Tags)
	if err != nil {
		return restic.BackupOptions{}, fmt.Errorf("plan %q: decode tags: %w", plan.Name, err)
	}
	out := restic.BackupOptions{
		Tags:              append([]string{plan.Tag}, tags...),
		Host:              repo.Host,
		ExcludeLargerThan: opts.ExcludeLargerThan,
		Compression:       opts.Compression,
		OneFileSystem:     opts.OneFileSystem,
		ExcludeCaches:     opts.ExcludeCaches,
		ExcludeIfPresent:  opts.ExcludeIfPresent,
		LimitUpload:       opts.LimitUpload,
		LimitDownload:     opts.LimitDownload,
	}
	for _, p := range paths {
		if p.Kind == "exclude" {
			out.Excludes = append(out.Excludes, p.Path)
		} else {
			out.Paths = append(out.Paths, p.Path)
		}
	}
	return out, nil
}

func (m *Manager) backup(ctx context.Context, job db.Job, repo db.Repository, plan db.Plan, target *restic.Repo) outcome {
	paths, err := m.store.ListPlanPaths(ctx, plan.ID)
	if err != nil {
		return failed(err)
	}
	opts, err := BackupOptions(repo, plan, paths)
	if err != nil {
		return failed(err)
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

func (m *Manager) forget(ctx context.Context, repo db.Repository, plan db.Plan, target *restic.Repo) outcome {
	ret, err := m.store.GetPlanRetention(ctx, plan.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return failed(ErrNoRetention)
	}
	if err != nil {
		return failed(err)
	}
	opts := ForgetOptions(repo, plan, restic.Policy{
		KeepLast: int(ret.KeepLast), KeepHourly: int(ret.KeepHourly), KeepDaily: int(ret.KeepDaily),
		KeepWeekly: int(ret.KeepWeekly), KeepMonthly: int(ret.KeepMonthly), KeepYearly: int(ret.KeepYearly),
		KeepWithin: ret.KeepWithin,
	})
	opts.Prune = ret.Prune
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

func (m *Manager) restore(ctx context.Context, job db.Job, target *restic.Repo) outcome {
	var p RestoreParams
	if err := json.Unmarshal([]byte(job.Params), &p); err != nil {
		return failed(fmt.Errorf("restore parameters: %w", err))
	}
	var last time.Time
	summary, err := m.restic.Restore(ctx, target, restic.RestoreOptions{SnapshotID: p.SnapshotID, Target: p.Target, Include: p.Include}, func(s restic.RestoreStatus) {
		if now := time.Now(); now.Sub(last) >= m.progressInterval {
			last = now
			m.publish(job, EventProgress, StatusRunning, &Progress{
				PercentDone: s.PercentDone, FilesDone: s.FilesRestored, TotalFiles: s.TotalFiles,
				BytesDone: s.BytesRestored, TotalBytes: s.TotalBytes, SecondsRemaining: s.SecondsRemaining,
			})
		}
	})
	if err != nil {
		return failed(err)
	}
	return outcome{status: StatusSucceeded, summary: summary}
}

// ForgetOptions applies policy to the snapshots of a plan, in one group, so
// that editing the plan's paths or tags does not start a new retention
// group, and the other plans' snapshots are left alone. The retention
// preview uses it too, so that it shows what a forget job would do.
func ForgetOptions(repo db.Repository, plan db.Plan, policy restic.Policy) restic.ForgetOptions {
	return restic.ForgetOptions{
		Policy:  policy,
		Filter:  restic.SnapshotFilter{Host: repo.Host, Tags: []string{plan.Tag}},
		GroupBy: "host",
	}
}

// RepairSummary is the summary recorded for a repair_snapshots job.
type RepairSummary struct {
	// Modified is how many snapshots were rewritten or removed.
	Modified int `json:"modified"`
}

// formatV2 is the migration from repository format version 1 to 2, which
// brings compression. It is the only one ResticGUI applies.
const formatV2 = "upgrade_repo_v2"

// MigrateSummary is the summary recorded for a migrate job.
type MigrateSummary struct {
	// Applied lists the migrations applied, empty when the repository
	// already had the latest format.
	Applied []string `json:"applied"`
}

func (m *Manager) migrate(ctx context.Context, target *restic.Repo) outcome {
	available, err := m.restic.Migrations(ctx, target)
	if err != nil {
		return failed(err)
	}
	summary := MigrateSummary{Applied: []string{}}
	if slices.Contains(available, formatV2) {
		if err := m.restic.Migrate(ctx, target, formatV2); err != nil {
			return failed(err)
		}
		summary.Applied = append(summary.Applied, formatV2)
	}
	return outcome{status: StatusSucceeded, summary: summary}
}

// changePassword adds a key for password, saves password as the
// repository's, then removes the key of the old password. Whatever step
// fails, the saved password opens the repository.
func (m *Manager) changePassword(ctx context.Context, repo db.Repository, target *restic.Repo, password string) outcome {
	keys, err := m.restic.Keys(ctx, target)
	if err != nil {
		return failed(err)
	}
	old := currentKey(keys)
	if old == "" {
		return failed(errors.New("restic did not say which key the current password opens"))
	}
	if err := m.restic.AddKey(ctx, target, password); err != nil {
		return failed(err)
	}
	next := *target
	next.Password = password

	if err := m.savePassword(ctx, repo, password); err != nil {
		// the old password goes on opening the repository: take the new key away
		if keys, kerr := m.restic.Keys(ctx, &next); kerr == nil && currentKey(keys) != "" {
			_ = m.restic.RemoveKey(ctx, target, currentKey(keys))
		}
		return failed(fmt.Errorf("could not save the new password, the old one is kept: %w", err))
	}
	if err := m.restic.RemoveKey(ctx, &next, old); err != nil {
		out := failed(fmt.Errorf("the new password is saved, but the old one still opens the repository: %w", err))
		out.status = StatusSucceeded
		return out
	}
	return outcome{status: StatusSucceeded}
}

func currentKey(keys []restic.Key) string {
	for _, k := range keys {
		if k.Current {
			return k.ID
		}
	}
	return ""
}

func (m *Manager) savePassword(ctx context.Context, repo db.Repository, password string) error {
	sec, err := store.OpenSecrets(m.box, repo.Secrets)
	if err != nil {
		return err
	}
	sec.Password = password
	sealed, err := store.SealSecrets(m.box, sec)
	if err != nil {
		return err
	}
	n, err := m.store.UpdateRepositorySecrets(ctx, db.UpdateRepositorySecretsParams{ID: repo.ID, Secrets: sealed})
	if err == nil && n == 0 {
		err = ErrRepositoryNotFound
	}
	return err
}
