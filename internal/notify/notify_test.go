package notify

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// receiver is a local HTTP server that records what it is sent.
type receiver struct {
	*httptest.Server
	mu   sync.Mutex
	reqs []request
}

type request struct {
	Method, Path, Body string
	Header             http.Header
}

func newReceiver(t *testing.T) *receiver {
	r := &receiver{}
	r.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		r.mu.Lock()
		r.reqs = append(r.reqs, request{Method: req.Method, Path: req.URL.Path, Body: string(body), Header: req.Header})
		r.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","event":"message"}`))
	}))
	t.Cleanup(r.Close)
	return r
}

func (r *receiver) requests() []request {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]request(nil), r.reqs...)
}

func TestShoutrrrReachesNtfy(t *testing.T) {
	rec := newReceiver(t)
	host := strings.TrimPrefix(rec.URL, "http://")
	require.NoError(t, Shoutrrr("ntfy://"+host+"/backups?scheme=http", "❌ Backup of Photos failed", "repository is locked"))
	reqs := rec.requests()
	require.Len(t, reqs, 1)
	assert.Equal(t, "/backups", reqs[0].Path)
	assert.Contains(t, reqs[0].Body, "repository is locked")
	assert.Contains(t, reqs[0].Header.Get("Title")+reqs[0].Body, "Backup of Photos failed")
}

func TestShoutrrrReachesAGenericWebhook(t *testing.T) {
	rec := newReceiver(t)
	host := strings.TrimPrefix(rec.URL, "http://")
	require.NoError(t, Shoutrrr("generic://"+host+"/hook?disabletls=yes&template=json", "title", "hello"))
	reqs := rec.requests()
	require.Len(t, reqs, 1)
	assert.Equal(t, "/hook", reqs[0].Path)
	assert.Contains(t, reqs[0].Body, "hello")
}

func TestServiceName(t *testing.T) {
	name, err := ServiceName("ntfy://ntfy.sh/my-backups")
	require.NoError(t, err)
	assert.Equal(t, "ntfy", name)
	name, err = ServiceName("generic+https://example.com/hook")
	require.NoError(t, err)
	assert.Equal(t, "generic", name)

	_, err = ServiceName("nosuchservice://token-abcdef@host")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "token-abcdef")
	_, err = ServiceName("just text")
	assert.Error(t, err)
}

func TestRedact(t *testing.T) {
	raw := "discord://SeCrEtToKeN@123456789012/path?apikey=hunter2hunter2"
	err := Redact(errors.New(`post to `+raw+` failed: token SeCrEtToKeN rejected, key hunter2hunter2`), raw)
	assert.Equal(t, "post to … failed: token … rejected, key …", err.Error())
}

type env struct {
	store *store.Store
	box   *secrets.Box
	svc   *Service
	sent  *[]string
}

func newEnv(t *testing.T) *env {
	t.Helper()
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "resticgui.db"))
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })
	box, err := secrets.New(bytes.Repeat([]byte{7}, secrets.KeySize))
	require.NoError(t, err)
	var mu sync.Mutex
	sent := []string{}
	svc := &Service{Store: st, Box: box, Events: jobs.NewBroker(), Send: func(rawURL, title, body string) error {
		mu.Lock()
		defer mu.Unlock()
		sent = append(sent, rawURL+" | "+title+" | "+body)
		return nil
	}}
	return &env{store: st, box: box, svc: svc, sent: &sent}
}

func (e *env) repository(t *testing.T, name string) db.Repository {
	t.Helper()
	sealed, err := store.SealSecrets(e.box, store.RepositorySecrets{Password: "pw"})
	require.NoError(t, err)
	r, err := e.store.CreateRepository(context.Background(), db.CreateRepositoryParams{
		Name: name, Backend: "local", Location: "/srv/" + name, Options: "{}", Secrets: sealed,
	})
	require.NoError(t, err)
	e.plan(t, r.ID, "Default")
	return r
}

func (e *env) plan(t *testing.T, repositoryID int64, name string) db.Plan {
	t.Helper()
	var plan db.Plan
	require.NoError(t, e.store.InTx(context.Background(), func(q *db.Queries) error {
		var err error
		plan, err = store.SavePlan(context.Background(), q, repositoryID, 0, store.Plan{Name: name, Include: []string{"/data"}})
		return err
	}))
	return plan
}

// firstPlan is the plan made with the repository.
func (e *env) firstPlan(t *testing.T, repositoryID int64) db.Plan {
	t.Helper()
	plans, err := e.store.ListPlans(context.Background(), repositoryID)
	require.NoError(t, err)
	return plans[0]
}

func (e *env) notifier(t *testing.T, name, rawURL string, failure, success, missed, enabled bool) {
	t.Helper()
	sealed, err := e.box.Seal([]byte(rawURL))
	require.NoError(t, err)
	_, err = e.store.CreateNotifier(context.Background(), db.CreateNotifierParams{
		Name: name, Url: sealed, Service: "ntfy", OnFailure: failure, OnSuccess: success, OnMissed: missed, Enabled: enabled,
	})
	require.NoError(t, err)
}

func (e *env) finishedJob(t *testing.T, repoID int64, kind jobs.Kind, status jobs.Status, summary, errText string) db.Job {
	t.Helper()
	ctx := context.Background()
	var plan *int64
	if kind == jobs.KindBackup || kind == jobs.KindForget {
		id := e.firstPlan(t, repoID).ID
		plan = &id
	}
	job, err := e.store.CreateJob(ctx, db.CreateJobParams{RepositoryID: repoID, PlanID: plan, Kind: string(kind), Trigger: "schedule", Params: "{}"})
	require.NoError(t, err)
	_, err = e.store.FinishJob(ctx, db.FinishJobParams{ID: job.ID, Status: string(status), Summary: summary, Error: errText})
	require.NoError(t, err)
	job, err = e.store.GetJob(ctx, job.ID)
	require.NoError(t, err)
	return job
}

func TestAFailedBackupIsSentToTheNotifiersThatWantIt(t *testing.T) {
	e := newEnv(t)
	repo := e.repository(t, "Photos")
	e.notifier(t, "phone", "ntfy://ntfy.sh/phone", true, false, true, true)
	e.notifier(t, "successes only", "ntfy://ntfy.sh/ok", false, true, false, true)
	e.notifier(t, "paused", "ntfy://ntfy.sh/paused", true, true, true, false)
	job := e.finishedJob(t, repo.ID, jobs.KindBackup, jobs.StatusFailed, "{}", "repository is locked")

	e.svc.Handle(context.Background(), jobs.Event{Type: jobs.EventStatus, JobID: job.ID, RepositoryID: repo.ID, Kind: jobs.KindBackup, Status: jobs.StatusFailed})
	assert.Equal(t, []string{"ntfy://ntfy.sh/phone | ❌ Backup of Photos failed | Backup of Photos failed: repository is locked."}, *e.sent)
}

func TestASuccessSaysWhatWasBackedUp(t *testing.T) {
	e := newEnv(t)
	repo := e.repository(t, "Photos")
	e.notifier(t, "phone", "ntfy://ntfy.sh/phone", true, true, true, true)
	job := e.finishedJob(t, repo.ID, jobs.KindBackup, jobs.StatusSucceeded,
		`{"snapshot_id":"f0feba17aa","files_new":3,"files_changed":1,"data_added":1572864,"total_duration":4.2}`, "")

	e.svc.Handle(context.Background(), jobs.Event{Type: jobs.EventStatus, JobID: job.ID, RepositoryID: repo.ID, Kind: jobs.KindBackup, Status: jobs.StatusSucceeded})
	assert.Equal(t, []string{"ntfy://ntfy.sh/phone | ✅ Backup of Photos succeeded | Backup of Photos succeeded: snapshot f0feba17, 3 new and 1 changed files, 1.5 MiB added in 4s."}, *e.sent)
}

func TestRestoresAndQueuedJobsAreNotNotified(t *testing.T) {
	e := newEnv(t)
	repo := e.repository(t, "Photos")
	e.notifier(t, "phone", "ntfy://ntfy.sh/phone", true, true, true, true)
	job := e.finishedJob(t, repo.ID, jobs.KindRestore, jobs.StatusFailed, "{}", "no space left")
	e.svc.Handle(context.Background(), jobs.Event{Type: jobs.EventStatus, JobID: job.ID, RepositoryID: repo.ID, Kind: jobs.KindRestore, Status: jobs.StatusFailed})
	e.svc.Handle(context.Background(), jobs.Event{Type: jobs.EventStatus, JobID: job.ID, RepositoryID: repo.ID, Kind: jobs.KindBackup, Status: jobs.StatusQueued})
	assert.Empty(t, *e.sent)
}

func TestHealthchecksPings(t *testing.T) {
	e := newEnv(t)
	repo := e.repository(t, "Photos")
	rec := newReceiver(t)
	sealed, err := e.box.Seal([]byte(rec.URL + "/ping/abc-123/"))
	require.NoError(t, err)
	require.NoError(t, e.store.UpsertHealthcheck(context.Background(), db.UpsertHealthcheckParams{RepositoryID: repo.ID, Url: sealed, Host: "127.0.0.1"}))
	failedJob := e.finishedJob(t, repo.ID, jobs.KindBackup, jobs.StatusFailed, "{}", "locked")

	ctx := context.Background()
	e.svc.Handle(ctx, jobs.Event{JobID: failedJob.ID, RepositoryID: repo.ID, Kind: jobs.KindBackup, Status: jobs.StatusRunning})
	e.svc.Handle(ctx, jobs.Event{JobID: failedJob.ID, RepositoryID: repo.ID, Kind: jobs.KindBackup, Status: jobs.StatusSucceeded, SnapshotID: "f0feba17"})
	e.svc.Handle(ctx, jobs.Event{JobID: failedJob.ID, RepositoryID: repo.ID, Kind: jobs.KindBackup, Status: jobs.StatusFailed, Error: "locked"})
	e.svc.Handle(ctx, jobs.Event{JobID: failedJob.ID, RepositoryID: repo.ID, Kind: jobs.KindCheck, Status: jobs.StatusFailed})

	reqs := rec.requests()
	require.Len(t, reqs, 3, "only backups ping")
	assert.Equal(t, "/ping/abc-123/start", reqs[0].Path)
	assert.Equal(t, "/ping/abc-123", reqs[1].Path)
	assert.Equal(t, "Backup succeeded, snapshot f0feba17", reqs[1].Body)
	assert.Equal(t, "/ping/abc-123/fail", reqs[2].Path)
	assert.Equal(t, "Backup failed: locked", reqs[2].Body)
}

func TestMissedBackupsAreReportedOnce(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	repo := e.repository(t, "Photos")
	e.notifier(t, "phone", "ntfy://ntfy.sh/phone", true, false, true, true)
	require.NoError(t, e.store.UpsertPlanSchedule(ctx, db.UpsertPlanScheduleParams{PlanID: e.firstPlan(t, repo.ID).ID, Job: "backup", Cron: "30 2 * * *", Enabled: true}))

	// on time: nothing is due yet
	e.svc.Now = func() time.Time { return repo.CreatedAt.Add(time.Minute) }
	e.svc.CheckMissed(ctx)
	assert.Empty(t, *e.sent)

	// two days later, no backup ran
	e.svc.Now = func() time.Time { return repo.CreatedAt.Add(48 * time.Hour) }
	e.svc.CheckMissed(ctx)
	require.Len(t, *e.sent, 1)
	assert.Contains(t, (*e.sent)[0], "⏰ Backup of Photos did not run")
	e.svc.CheckMissed(ctx)
	assert.Len(t, *e.sent, 1, "reported once")

	// a busy repository is not late, it is running
	e.svc.Busy = func(int64) bool { return true }
	e.svc.Now = func() time.Time { return repo.CreatedAt.Add(96 * time.Hour) }
	e.svc.CheckMissed(ctx)
	assert.Len(t, *e.sent, 1)
}

func TestAFailedRunIsNotAlsoMissed(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	repo := e.repository(t, "Photos")
	e.notifier(t, "phone", "ntfy://ntfy.sh/phone", true, false, true, true)
	require.NoError(t, e.store.UpsertPlanSchedule(ctx, db.UpsertPlanScheduleParams{PlanID: e.firstPlan(t, repo.ID).ID, Job: "backup", Cron: "@every 1h", Enabled: true}))
	job := e.finishedJob(t, repo.ID, jobs.KindBackup, jobs.StatusFailed, "{}", "locked")

	// the failure was already reported; within the hour after it, nothing is missed
	e.svc.Now = func() time.Time { return job.FinishedAt.Add(90 * time.Minute) }
	e.svc.CheckMissed(ctx)
	assert.Empty(t, *e.sent)
	e.svc.Now = func() time.Time { return job.FinishedAt.Add(3 * time.Hour) }
	e.svc.CheckMissed(ctx)
	assert.Len(t, *e.sent, 1)
}

// With several plans, messages say which plan, and each plan's schedule is
// watched on its own.
func TestPlansAreNamedOnceThereAreSeveral(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	repo := e.repository(t, "Home")
	e.notifier(t, "phone", "ntfy://ntfy.sh/phone", true, false, true, true)
	first := e.firstPlan(t, repo.ID)
	videos := e.plan(t, repo.ID, "Videos")
	for _, p := range []db.Plan{first, videos} {
		require.NoError(t, e.store.UpsertPlanSchedule(ctx, db.UpsertPlanScheduleParams{PlanID: p.ID, Job: "backup", Cron: "@every 1h", Enabled: true}))
	}
	job := e.finishedJob(t, repo.ID, jobs.KindBackup, jobs.StatusFailed, "{}", "locked")
	e.svc.Handle(ctx, jobs.Event{Type: jobs.EventStatus, JobID: job.ID, RepositoryID: repo.ID, Kind: jobs.KindBackup, Status: jobs.StatusFailed})
	require.Len(t, *e.sent, 1)
	assert.Contains(t, (*e.sent)[0], "❌ Backup of Home (Default) failed")

	// later, the first plan backed up again, so only the other one is late
	later := job.FinishedAt.Add(3 * time.Hour)
	_, err := e.store.DB.ExecContext(ctx, `UPDATE jobs SET finished_at = ? WHERE id = ?`, later, job.ID)
	require.NoError(t, err)
	e.svc.Now = func() time.Time { return later.Add(time.Minute) }
	e.svc.CheckMissed(ctx)
	require.Len(t, *e.sent, 2)
	assert.Contains(t, (*e.sent)[1], "⏰ Backup of Home (Videos) did not run")
}

func TestMissed(t *testing.T) {
	due := time.Date(2026, 10, 18, 2, 30, 0, 0, time.UTC)
	assert.False(t, Missed(due, due.Add(Grace)))
	assert.True(t, Missed(due, due.Add(Grace+time.Second)))
	assert.False(t, Missed(time.Time{}, due))
}

func TestBytes(t *testing.T) {
	assert.Equal(t, "512 B", Bytes(512))
	assert.Equal(t, "1.5 KiB", Bytes(1536))
	assert.Equal(t, "2.0 GiB", Bytes(2<<30))
}
