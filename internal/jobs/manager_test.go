package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// fakeCall is one call the manager made to restic.
type fakeCall struct {
	Kind string
	Repo *restic.Repo
	Opts any
}

// fakeRestic records calls. When run is set, every call goes through it,
// which lets a test block a job or make it fail.
type fakeRestic struct {
	mu       sync.Mutex
	calls    []fakeCall
	run      func(ctx context.Context, c fakeCall) error
	progress []restic.BackupStatus
	backup   *restic.BackupResult
	groups   []restic.ForgetGroup
	check    *restic.CheckSummary
}

func (f *fakeRestic) do(ctx context.Context, kind string, repo *restic.Repo, opts any) error {
	c := fakeCall{Kind: kind, Repo: repo, Opts: opts}
	f.mu.Lock()
	f.calls = append(f.calls, c)
	run := f.run
	f.mu.Unlock()
	if run != nil {
		return run(ctx, c)
	}
	return nil
}

func (f *fakeRestic) Calls() []fakeCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]fakeCall(nil), f.calls...)
}

func (f *fakeRestic) Init(ctx context.Context, repo *restic.Repo) error {
	return f.do(ctx, "init", repo, nil)
}

func (f *fakeRestic) Unlock(ctx context.Context, repo *restic.Repo) error {
	return f.do(ctx, "unlock", repo, nil)
}

func (f *fakeRestic) Backup(ctx context.Context, repo *restic.Repo, opts restic.BackupOptions, onProgress func(restic.BackupStatus)) (*restic.BackupResult, error) {
	for _, s := range f.progress {
		onProgress(s)
	}
	err := f.do(ctx, "backup", repo, opts)
	if err != nil && !errors.Is(err, restic.ErrIncomplete) {
		return nil, err
	}
	return f.backup, err
}

func (f *fakeRestic) Forget(ctx context.Context, repo *restic.Repo, opts restic.ForgetOptions) ([]restic.ForgetGroup, error) {
	return f.groups, f.do(ctx, "forget", repo, opts)
}

func (f *fakeRestic) Prune(ctx context.Context, repo *restic.Repo) error {
	return f.do(ctx, "prune", repo, nil)
}

func (f *fakeRestic) Check(ctx context.Context, repo *restic.Repo, opts restic.CheckOptions) (*restic.CheckSummary, error) {
	return f.check, f.do(ctx, "check", repo, opts)
}

func (f *fakeRestic) Restore(ctx context.Context, repo *restic.Repo, opts restic.RestoreOptions, onProgress func(restic.RestoreStatus)) (*restic.RestoreSummary, error) {
	onProgress(restic.RestoreStatus{PercentDone: 0.25, FilesRestored: 1, TotalFiles: 4})
	if err := f.do(ctx, "restore", repo, opts); err != nil {
		return nil, err
	}
	return &restic.RestoreSummary{FilesRestored: 4, TotalFiles: 4}, nil
}

// gate blocks the restic calls of a test until released, and reports which
// repositories are running.
type gate struct {
	started chan string
	release chan struct{}
}

func newGate() *gate {
	return &gate{started: make(chan string, 16), release: make(chan struct{})}
}

func (g *gate) run(ctx context.Context, c fakeCall) error {
	g.started <- c.Repo.Location
	select {
	case <-g.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *gate) waitStarted(t *testing.T) string {
	t.Helper()
	select {
	case loc := <-g.started:
		return loc
	case <-time.After(5 * time.Second):
		t.Fatal("no restic call started")
		return ""
	}
}

func (g *gate) assertNoStart(t *testing.T) {
	t.Helper()
	select {
	case loc := <-g.started:
		t.Fatalf("unexpected restic call on %s", loc)
	case <-time.After(100 * time.Millisecond):
	}
}

type env struct {
	store   *store.Store
	box     *secrets.Box
	restic  *fakeRestic
	events  *Broker
	manager *Manager
}

func newEnv(t *testing.T, opts Options) *env {
	t.Helper()
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "resticgui.db"))
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })
	box, err := secrets.New(bytes.Repeat([]byte{7}, secrets.KeySize))
	require.NoError(t, err)
	e := &env{store: st, box: box, restic: &fakeRestic{}, events: NewBroker()}
	if opts.ProgressInterval == 0 {
		opts.ProgressInterval = time.Nanosecond
	}
	e.manager = New(st, box, e.restic, e.events, opts)
	t.Cleanup(e.manager.Close)
	return e
}

func (e *env) repository(t *testing.T, name string) db.Repository {
	t.Helper()
	sealed, err := store.SealSecrets(e.box, store.RepositorySecrets{
		Password: "pw-" + name,
		Env:      map[string]string{"AWS_SECRET_ACCESS_KEY": "secret-" + name},
	})
	require.NoError(t, err)
	options, err := store.EncodeOptions(store.RepositoryOptions{
		Env:      map[string]string{"AWS_DEFAULT_REGION": "eu-west-1"},
		CACert:   "/etc/ca.pem",
		Extended: map[string]string{"s3.storage-class": "STANDARD_IA"},
	})
	require.NoError(t, err)
	r, err := e.store.CreateRepository(context.Background(), db.CreateRepositoryParams{
		Name: name, Backend: "s3", Location: "s3:example.com/" + name, Options: options,
		Secrets: sealed, Host: "nas", Compression: "max", ExcludeLargerThan: "1G",
	})
	require.NoError(t, err)
	return r
}

func (e *env) waitStatus(t *testing.T, jobID int64, want Status) db.Job {
	t.Helper()
	var job db.Job
	require.Eventually(t, func() bool {
		var err error
		job, err = e.store.GetJob(context.Background(), jobID)
		require.NoError(t, err)
		return Status(job.Status) == want
	}, 5*time.Second, 5*time.Millisecond, "job %d never became %s (last: %s)", jobID, want, job.Status)
	return job
}

func TestBackupJob(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "photos")
	for _, p := range []db.AddPathParams{
		{RepositoryID: repo.ID, Path: "/home/me/Pictures", Kind: "include"},
		{RepositoryID: repo.ID, Path: "/home/me/Pictures/.cache", Kind: "exclude"},
		{RepositoryID: repo.ID, Path: "/home/me/Documents", Kind: "include"},
	} {
		_, err := e.store.AddPath(ctx, p)
		require.NoError(t, err)
	}
	e.restic.progress = []restic.BackupStatus{{PercentDone: 0.5, FilesDone: 1, TotalFiles: 2, BytesDone: 10, TotalBytes: 20}}
	e.restic.backup = &restic.BackupResult{Summary: &restic.BackupSummary{SnapshotID: "abc123", TotalDuration: 1.5}}
	events, unsubscribe := e.events.Subscribe(16)
	defer unsubscribe()

	job, err := e.manager.Enqueue(ctx, repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	assert.Equal(t, "queued", job.Status)
	assert.Equal(t, "manual", job.Trigger)

	done := e.waitStatus(t, job.ID, StatusSucceeded)
	assert.Equal(t, "abc123", done.SnapshotID)
	var summary restic.BackupSummary
	require.NoError(t, json.Unmarshal([]byte(done.Summary), &summary))
	assert.Equal(t, *e.restic.backup.Summary, summary)
	assert.NotNil(t, done.StartedAt)
	assert.NotNil(t, done.FinishedAt)
	assert.Empty(t, done.Error)

	calls := e.restic.Calls()
	require.Len(t, calls, 1)
	assert.Equal(t, &restic.Repo{
		Location: "s3:example.com/photos",
		Password: "pw-photos",
		Env:      map[string]string{"AWS_DEFAULT_REGION": "eu-west-1", "AWS_SECRET_ACCESS_KEY": "secret-photos"},
		CACert:   "/etc/ca.pem",
		Options:  map[string]string{"s3.storage-class": "STANDARD_IA"},
	}, calls[0].Repo)
	assert.Equal(t, restic.BackupOptions{
		Paths:             []string{"/home/me/Documents", "/home/me/Pictures"},
		Excludes:          []string{"/home/me/Pictures/.cache"},
		Tags:              []string{Tag(repo.ID)},
		Host:              "nas",
		ExcludeLargerThan: "1G",
		Compression:       "max",
	}, calls[0].Opts)

	var got []Event
	for len(got) < 4 {
		select {
		case ev := <-events:
			got = append(got, ev)
		case <-time.After(5 * time.Second):
			t.Fatalf("missing events, got %+v", got)
		}
	}
	base := Event{JobID: job.ID, RepositoryID: repo.ID, Kind: KindBackup}
	want := []Event{base, base, base, base}
	want[0].Type, want[0].Status = EventStatus, StatusQueued
	want[1].Type, want[1].Status = EventStatus, StatusRunning
	want[2].Type, want[2].Status = EventProgress, StatusRunning
	want[2].Progress = &Progress{PercentDone: 0.5, FilesDone: 1, TotalFiles: 2, BytesDone: 10, TotalBytes: 20}
	want[3].Type, want[3].Status, want[3].SnapshotID = EventStatus, StatusSucceeded, "abc123"
	assert.Equal(t, want, got)
}

func TestProgressIsThrottled(t *testing.T) {
	e := newEnv(t, Options{ProgressInterval: time.Hour})
	repo := e.repository(t, "r")
	_, err := e.store.AddPath(context.Background(), db.AddPathParams{RepositoryID: repo.ID, Path: "/data", Kind: "include"})
	require.NoError(t, err)
	e.restic.progress = make([]restic.BackupStatus, 50)
	events, unsubscribe := e.events.Subscribe(64)
	defer unsubscribe()

	job, err := e.manager.Enqueue(context.Background(), repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	e.waitStatus(t, job.ID, StatusSucceeded)
	progress := 0
	for ev := range drain(events) {
		if ev.Type == EventProgress {
			progress++
		}
	}
	assert.Equal(t, 1, progress, "only the first status within the interval is published")
}

func drain(ch <-chan Event) <-chan Event {
	out := make(chan Event, cap(ch))
	defer close(out)
	for {
		select {
		case ev := <-ch:
			out <- ev
		case <-time.After(50 * time.Millisecond):
			return out
		}
	}
}

func TestIncompleteBackupSucceedsWithError(t *testing.T) {
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	_, err := e.store.AddPath(context.Background(), db.AddPathParams{RepositoryID: repo.ID, Path: "/data", Kind: "include"})
	require.NoError(t, err)
	e.restic.backup = &restic.BackupResult{
		Summary: &restic.BackupSummary{SnapshotID: "partial1"},
		Errors:  []restic.BackupError{{Item: "/data/locked.db", Message: "permission denied"}},
	}
	e.restic.run = func(context.Context, fakeCall) error {
		return &restic.Error{Command: "backup", ExitCode: 3}
	}

	job, err := e.manager.Enqueue(context.Background(), repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusSucceeded)
	assert.Equal(t, "partial1", done.SnapshotID)
	assert.Equal(t, "1 files could not be read, first: /data/locked.db: permission denied", done.Error)
	require.NotNil(t, done.ExitCode)
	assert.EqualValues(t, 3, *done.ExitCode)
}

func TestFailedJobRecordsResticError(t *testing.T) {
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	e.restic.run = func(context.Context, fakeCall) error {
		return &restic.Error{Command: "check", ExitCode: 12, Message: "wrong password or no key found", Stderr: []string{"line 1", "wrong password or no key found"}}
	}

	job, err := e.manager.Enqueue(context.Background(), repo.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusFailed)
	assert.Equal(t, "restic check: exit code 12: wrong password or no key found", done.Error)
	require.NotNil(t, done.ExitCode)
	assert.EqualValues(t, 12, *done.ExitCode)
	assert.Equal(t, "line 1\nwrong password or no key found", done.LogExcerpt)
}

func TestForgetJobUsesRetention(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")

	job, err := e.manager.Enqueue(ctx, repo.ID, KindForget, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusFailed)
	assert.Equal(t, ErrNoRetention.Error(), done.Error)
	assert.Empty(t, e.restic.Calls(), "restic is not called without a policy")

	_, err = e.store.UpsertRetention(ctx, db.UpsertRetentionParams{RepositoryID: repo.ID, KeepDaily: 7, KeepYearly: 2, KeepWithin: "30d", Prune: true})
	require.NoError(t, err)
	e.restic.groups = []restic.ForgetGroup{
		{Keep: make([]restic.Snapshot, 3), Remove: make([]restic.Snapshot, 2)},
		{Keep: make([]restic.Snapshot, 1)},
	}
	job, err = e.manager.Enqueue(ctx, repo.ID, KindForget, TriggerSchedule)
	require.NoError(t, err)
	done = e.waitStatus(t, job.ID, StatusSucceeded)
	assert.JSONEq(t, `{"kept":4,"removed":2,"pruned":true}`, done.Summary)

	calls := e.restic.Calls()
	require.Len(t, calls, 1)
	assert.Equal(t, restic.ForgetOptions{
		Policy:  restic.Policy{KeepDaily: 7, KeepYearly: 2, KeepWithin: "30d"},
		Filter:  restic.SnapshotFilter{Host: "nas", Tags: []string{Tag(repo.ID)}},
		GroupBy: "host,tags",
		Prune:   true,
	}, calls[0].Opts)
}

func TestInitJobMarksRepositoryInitialized(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	require.Nil(t, repo.InitializedAt)

	job, err := e.manager.Enqueue(ctx, repo.ID, KindInit, TriggerManual)
	require.NoError(t, err)
	e.waitStatus(t, job.ID, StatusSucceeded)
	repo, err = e.store.GetRepository(ctx, repo.ID)
	require.NoError(t, err)
	assert.NotNil(t, repo.InitializedAt)
}

func TestSucceededJobsMarkRepositoryInitialized(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	initialized := func(id int64) bool {
		r, err := e.store.GetRepository(ctx, id)
		require.NoError(t, err)
		return r.InitializedAt != nil
	}
	run := func(id int64, kind Kind, want Status) {
		job, err := e.manager.Enqueue(ctx, id, kind, TriggerManual)
		require.NoError(t, err)
		e.waitStatus(t, job.ID, want)
	}

	// a repository made with restic on the command line never has an init job
	existing := e.repository(t, "existing")
	run(existing.ID, KindUnlock, StatusSucceeded)
	assert.False(t, initialized(existing.ID), "unlock works on a missing repository too")
	run(existing.ID, KindCheck, StatusSucceeded)
	assert.True(t, initialized(existing.ID))

	missing := e.repository(t, "missing")
	e.restic.mu.Lock()
	e.restic.run = func(context.Context, fakeCall) error { return restic.ErrRepositoryNotFound }
	e.restic.mu.Unlock()
	run(missing.ID, KindBackup, StatusFailed)
	assert.False(t, initialized(missing.ID))
}

func TestOtherKinds(t *testing.T) {
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	e.restic.check = &restic.CheckSummary{NumErrors: 0}
	for _, kind := range []Kind{KindPrune, KindCheck, KindUnlock} {
		job, err := e.manager.Enqueue(context.Background(), repo.ID, kind, TriggerManual)
		require.NoError(t, err)
		e.waitStatus(t, job.ID, StatusSucceeded)
	}
	var kinds []string
	for _, c := range e.restic.Calls() {
		kinds = append(kinds, c.Kind)
	}
	assert.Equal(t, []string{"prune", "check", "unlock"}, kinds)
}

func TestEnqueueErrors(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")

	_, err := e.manager.Enqueue(ctx, repo.ID, "restore", TriggerManual)
	assert.ErrorIs(t, err, ErrUnknownKind)
	_, err = e.manager.Enqueue(ctx, 999, KindBackup, TriggerManual)
	assert.ErrorIs(t, err, ErrRepositoryNotFound)

	e.manager.Close()
	_, err = e.manager.Enqueue(ctx, repo.ID, KindCheck, TriggerManual)
	assert.ErrorIs(t, err, ErrClosed)
}

func TestJobsOfOneRepositoryRunInOrder(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{Workers: 4})
	g := newGate()
	e.restic.run = g.run
	repo := e.repository(t, "r")

	first, err := e.manager.Enqueue(ctx, repo.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	second, err := e.manager.Enqueue(ctx, repo.ID, KindPrune, TriggerManual)
	require.NoError(t, err)

	g.waitStarted(t)
	e.waitStatus(t, first.ID, StatusRunning)
	g.assertNoStart(t)
	job, err := e.store.GetJob(ctx, second.ID)
	require.NoError(t, err)
	assert.Equal(t, "queued", job.Status, "the second job waits for the repository")

	g.release <- struct{}{}
	e.waitStatus(t, first.ID, StatusSucceeded)
	g.waitStarted(t)
	e.waitStatus(t, second.ID, StatusRunning)
	g.release <- struct{}{}
	e.waitStatus(t, second.ID, StatusSucceeded)

	var kinds []string
	for _, c := range e.restic.Calls() {
		kinds = append(kinds, c.Kind)
	}
	assert.Equal(t, []string{"check", "prune"}, kinds)
}

func TestRepositoriesRunInParallelUpToWorkers(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{Workers: 2})
	g := newGate()
	e.restic.run = g.run
	a, b, c := e.repository(t, "a"), e.repository(t, "b"), e.repository(t, "c")

	jobA, err := e.manager.Enqueue(ctx, a.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	jobB, err := e.manager.Enqueue(ctx, b.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	started := []string{g.waitStarted(t), g.waitStarted(t)}
	assert.ElementsMatch(t, []string{a.Location, b.Location}, started, "two repositories run at once")

	jobC, err := e.manager.Enqueue(ctx, c.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	g.assertNoStart(t)

	g.release <- struct{}{}
	assert.Equal(t, c.Location, g.waitStarted(t), "a freed worker takes the next repository")
	g.release <- struct{}{}
	g.release <- struct{}{}
	for _, id := range []int64{jobA.ID, jobB.ID, jobC.ID} {
		e.waitStatus(t, id, StatusSucceeded)
	}
}

func TestScheduledJobSkippedWhileSameKindPending(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	g := newGate()
	e.restic.run = g.run
	repo := e.repository(t, "r")

	running, err := e.manager.Enqueue(ctx, repo.ID, KindCheck, TriggerSchedule)
	require.NoError(t, err)
	g.waitStarted(t)
	_, err = e.manager.Enqueue(ctx, repo.ID, KindCheck, TriggerSchedule)
	assert.ErrorIs(t, err, ErrAlreadyQueued)

	other, err := e.manager.Enqueue(ctx, repo.ID, KindPrune, TriggerSchedule)
	require.NoError(t, err, "another kind is queued")
	manual, err := e.manager.Enqueue(ctx, repo.ID, KindCheck, TriggerManual)
	require.NoError(t, err, "a manual run is always queued")

	for _, id := range []int64{running.ID, other.ID, manual.ID} {
		g.release <- struct{}{}
		e.waitStatus(t, id, StatusSucceeded)
		if id != manual.ID {
			g.waitStarted(t)
		}
	}
}

func TestCancelRunningJob(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	g := newGate()
	e.restic.run = g.run
	repo := e.repository(t, "r")

	job, err := e.manager.Enqueue(ctx, repo.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	g.waitStarted(t)
	require.NoError(t, e.manager.Cancel(ctx, job.ID))
	done := e.waitStatus(t, job.ID, StatusCancelled)
	assert.Empty(t, done.Error)
	assert.ErrorIs(t, e.manager.Cancel(ctx, job.ID), ErrFinished)
}

func TestCancelQueuedJob(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	g := newGate()
	e.restic.run = g.run
	repo := e.repository(t, "r")

	first, err := e.manager.Enqueue(ctx, repo.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	g.waitStarted(t)
	queued, err := e.manager.Enqueue(ctx, repo.ID, KindPrune, TriggerManual)
	require.NoError(t, err)
	require.NoError(t, e.manager.Cancel(ctx, queued.ID))
	e.waitStatus(t, queued.ID, StatusCancelled)

	g.release <- struct{}{}
	e.waitStatus(t, first.ID, StatusSucceeded)
	g.assertNoStart(t)
	assert.Len(t, e.restic.Calls(), 1, "the cancelled job never ran")
}

func TestCancelUnknownJob(t *testing.T) {
	e := newEnv(t, Options{})
	assert.ErrorIs(t, e.manager.Cancel(context.Background(), 42), ErrNotFound)
}

func TestCloseInterruptsJobs(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{Workers: 1})
	g := newGate()
	e.restic.run = g.run
	a, b := e.repository(t, "a"), e.repository(t, "b")

	running, err := e.manager.Enqueue(ctx, a.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	g.waitStarted(t)
	waitingForWorker, err := e.manager.Enqueue(ctx, b.ID, KindCheck, TriggerManual)
	require.NoError(t, err)
	queued, err := e.manager.Enqueue(ctx, a.ID, KindPrune, TriggerManual)
	require.NoError(t, err)

	e.manager.Close()
	for _, id := range []int64{running.ID, waitingForWorker.ID, queued.ID} {
		job, err := e.store.GetJob(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "interrupted", job.Status, "job %d", id)
	}
}

func TestRecoverMarksUnfinishedJobsInterrupted(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	// rows left behind by a process that died
	queued, err := e.store.CreateJob(ctx, db.CreateJobParams{RepositoryID: repo.ID, Kind: "backup", Trigger: "schedule", Params: "{}"})
	require.NoError(t, err)
	running, err := e.store.CreateJob(ctx, db.CreateJobParams{RepositoryID: repo.ID, Kind: "check", Trigger: "manual", Params: "{}"})
	require.NoError(t, err)
	_, err = e.store.StartJob(ctx, running.ID)
	require.NoError(t, err)

	n, err := e.manager.Recover(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 2, n)
	for _, id := range []int64{queued.ID, running.ID} {
		job, err := e.store.GetJob(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "interrupted", job.Status)
		assert.NotNil(t, job.FinishedAt)
	}
}

func TestBrokerDropsSlowSubscribers(t *testing.T) {
	b := NewBroker()
	ch, unsubscribe := b.Subscribe(1)
	defer unsubscribe()

	b.Publish(Event{Type: EventStatus, JobID: 1})
	b.Publish(Event{Type: EventProgress, JobID: 1}) // dropped, the buffer is full
	ev, ok := <-ch
	require.True(t, ok)
	assert.Equal(t, int64(1), ev.JobID)

	b.Publish(Event{Type: EventStatus, JobID: 2})
	b.Publish(Event{Type: EventStatus, JobID: 3}) // would be lost: the channel is closed
	<-ch
	_, ok = <-ch
	assert.False(t, ok, "a subscriber that misses a status event is closed")
	unsubscribe() // safe after the broker closed it
}

func TestRestoreJob(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	events, unsubscribe := e.events.Subscribe(16)
	defer unsubscribe()

	_, err := e.manager.Enqueue(ctx, repo.ID, KindRestore, TriggerManual)
	assert.ErrorIs(t, err, ErrUnknownKind, "a restore needs its parameters")
	_, err = e.manager.EnqueueRestore(ctx, repo.ID, RestoreParams{SnapshotID: "abc"})
	assert.Error(t, err)

	job, err := e.manager.EnqueueRestore(ctx, repo.ID, RestoreParams{SnapshotID: "abc", Target: "/restore", Include: []string{"/home/me/a.txt"}})
	require.NoError(t, err)
	assert.Equal(t, "restore", job.Kind)
	assert.JSONEq(t, `{"snapshot_id":"abc","target":"/restore","include":["/home/me/a.txt"]}`, job.Params)
	done := e.waitStatus(t, job.ID, StatusSucceeded)
	var summary restic.RestoreSummary
	require.NoError(t, json.Unmarshal([]byte(done.Summary), &summary))
	assert.EqualValues(t, 4, summary.FilesRestored)

	calls := e.restic.Calls()
	require.Len(t, calls, 1)
	assert.Equal(t, restic.RestoreOptions{SnapshotID: "abc", Target: "/restore", Include: []string{"/home/me/a.txt"}}, calls[0].Opts)
	var progress *Progress
	for ev := range drain(events) {
		if ev.Type == EventProgress {
			progress = ev.Progress
		}
	}
	assert.Equal(t, &Progress{PercentDone: 0.25, FilesDone: 1, TotalFiles: 4}, progress)
}
