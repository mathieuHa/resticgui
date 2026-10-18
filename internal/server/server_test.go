package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
)

// Values that must never appear in a response.
const (
	password  = "correct-horse-battery-staple"
	secretKey = "wJalrXUtnFEMI-K7MDENG-bPxRfiCY"
)

// fakeRestic serves snapshots and lets jobs block until released.
type fakeRestic struct {
	mu        sync.Mutex
	snapshots []restic.Snapshot
	err       error
	filters   []restic.SnapshotFilter
	block     chan struct{}
}

func (f *fakeRestic) Snapshots(_ context.Context, _ *restic.Repo, filter restic.SnapshotFilter) ([]restic.Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.filters = append(f.filters, filter)
	return f.snapshots, f.err
}

func (f *fakeRestic) wait(ctx context.Context) error {
	f.mu.Lock()
	block := f.block
	f.mu.Unlock()
	if block == nil {
		return nil
	}
	select {
	case <-block:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (f *fakeRestic) Init(ctx context.Context, _ *restic.Repo) error   { return f.wait(ctx) }
func (f *fakeRestic) Unlock(ctx context.Context, _ *restic.Repo) error { return f.wait(ctx) }
func (f *fakeRestic) Prune(ctx context.Context, _ *restic.Repo) error  { return f.wait(ctx) }
func (f *fakeRestic) Backup(ctx context.Context, _ *restic.Repo, _ restic.BackupOptions, _ func(restic.BackupStatus)) (*restic.BackupResult, error) {
	return &restic.BackupResult{Summary: &restic.BackupSummary{SnapshotID: "snap1"}}, f.wait(ctx)
}
func (f *fakeRestic) Forget(ctx context.Context, _ *restic.Repo, _ restic.ForgetOptions) ([]restic.ForgetGroup, error) {
	return nil, f.wait(ctx)
}
func (f *fakeRestic) Check(ctx context.Context, _ *restic.Repo, _ restic.CheckOptions) (*restic.CheckSummary, error) {
	return &restic.CheckSummary{}, f.wait(ctx)
}

type reloads struct{ n atomic.Int32 }

func (r *reloads) Reload() { r.n.Add(1) }

type api struct {
	t       *testing.T
	srv     *httptest.Server
	store   *store.Store
	restic  *fakeRestic
	reloads *reloads
}

func newAPI(t *testing.T) *api {
	t.Helper()
	gin.SetMode(gin.TestMode)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "resticgui.db"))
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })
	box, err := secrets.New(bytes.Repeat([]byte{9}, secrets.KeySize))
	require.NoError(t, err)
	fake := &fakeRestic{}
	manager := jobs.New(st, box, fake, jobs.NewBroker(), jobs.Options{})
	t.Cleanup(manager.Close)
	a := &api{t: t, store: st, restic: fake, reloads: &reloads{}}
	r := gin.New()
	(&Server{Store: st, Box: box, Restic: fake, Jobs: manager, Scheduler: a.reloads}).Register(r.Group(BasePath))
	a.srv = httptest.NewServer(r)
	t.Cleanup(a.srv.Close)
	return a
}

// do sends a request and checks that no secret leaks into the response.
func (a *api) do(method, path string, body any) (int, []byte) {
	a.t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(a.t, err)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, a.srv.URL+BasePath+path, reader)
	require.NoError(a.t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(a.t, err)
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	require.NoError(a.t, err)
	for _, secret := range []string{password, secretKey, "new-password", "rotated-key"} {
		require.NotContains(a.t, string(out), secret, "%s %s leaked a secret", method, path)
	}
	return resp.StatusCode, out
}

func (a *api) decode(body []byte, v any) {
	a.t.Helper()
	require.NoError(a.t, json.Unmarshal(body, v), string(body))
}

func s3Input(name string) map[string]any {
	return map[string]any{
		"name":     name,
		"backend":  "s3",
		"location": "s3:https://s3.example.com/bucket/" + name,
		"password": password,
		"env":      map[string]string{"AWS_DEFAULT_REGION": "eu-west-3", "AWS_ACCESS_KEY_ID": "AKIAEXAMPLE"},
		"secrets":  map[string]string{"AWS_SECRET_ACCESS_KEY": secretKey},
		"host":     "nas",
	}
}

func (a *api) create(name string) RepositoryDTO {
	a.t.Helper()
	status, body := a.do(http.MethodPost, "/repositories", s3Input(name))
	require.Equal(a.t, http.StatusCreated, status, string(body))
	var dto RepositoryDTO
	a.decode(body, &dto)
	return dto
}

func TestBackends(t *testing.T) {
	a := newAPI(t)
	status, body := a.do(http.MethodGet, "/backends", nil)
	require.Equal(t, http.StatusOK, status)
	var list []struct{ Name string }
	a.decode(body, &list)
	require.Len(t, list, 8)
	assert.Equal(t, "azure", list[0].Name)
}

func TestCreateAndGetRepository(t *testing.T) {
	a := newAPI(t)
	created := a.create("photos")
	assert.Equal(t, "photos", created.Name)
	assert.Equal(t, "s3", created.Backend)
	assert.Equal(t, map[string]string{"AWS_DEFAULT_REGION": "eu-west-3", "AWS_ACCESS_KEY_ID": "AKIAEXAMPLE"}, created.Env)
	assert.Equal(t, []string{"AWS_SECRET_ACCESS_KEY"}, created.Secrets)
	assert.True(t, created.HasPassword)
	assert.Equal(t, "auto", created.Compression)
	assert.Nil(t, created.InitializedAt)
	assert.Equal(t, PathsDTO{Include: []string{}, Exclude: []string{}}, created.Paths)
	assert.Nil(t, created.Retention)
	assert.Empty(t, created.Schedules)

	status, body := a.do(http.MethodGet, "/repositories/"+itoa(created.ID), nil)
	require.Equal(t, http.StatusOK, status)
	var got RepositoryDTO
	a.decode(body, &got)
	assert.Equal(t, created, got)

	status, body = a.do(http.MethodGet, "/repositories", nil)
	require.Equal(t, http.StatusOK, status)
	var list []RepositoryDTO
	a.decode(body, &list)
	assert.Equal(t, []RepositoryDTO{created}, list)

	status, _ = a.do(http.MethodPost, "/repositories", s3Input("photos"))
	assert.Equal(t, http.StatusConflict, status, "names are unique")
}

func TestCreateRepositoryValidation(t *testing.T) {
	a := newAPI(t)
	status, body := a.do(http.MethodPost, "/repositories", map[string]any{
		"backend": "s3", "location": "/not/s3", "compression": "fast",
		"env":                 map[string]string{"LD_PRELOAD": "/tmp/x.so"},
		"exclude_larger_than": "big",
		"extended":            map[string]string{"--no-lock": "true"},
	})
	require.Equal(t, http.StatusBadRequest, status)
	var e apiError
	a.decode(body, &e)
	details := strings.Join(e.Details, "\n")
	for _, want := range []string{
		"name is required", "password is required", `a s3 location starts with "s3:"`,
		"LD_PRELOAD", "AWS_SECRET_ACCESS_KEY is required", "compression is one of",
		"exclude_larger_than", `extended option "--no-lock"`,
	} {
		assert.Contains(t, details, want)
	}

	status, _ = a.do(http.MethodPost, "/repositories", "not an object")
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestUpdateRepositoryKeepsOrReplacesSecrets(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID)

	// no password and no secrets: both are kept
	in := s3Input("photos-renamed")
	delete(in, "password")
	delete(in, "secrets")
	in["compression"] = "max"
	status, body := a.do(http.MethodPut, path, in)
	require.Equal(t, http.StatusOK, status, string(body))
	var updated RepositoryDTO
	a.decode(body, &updated)
	assert.Equal(t, "photos-renamed", updated.Name)
	assert.Equal(t, "max", updated.Compression)
	assert.Equal(t, []string{"AWS_SECRET_ACCESS_KEY"}, updated.Secrets)
	sec := a.secrets(repo.ID)
	assert.Equal(t, password, sec.Password)
	assert.Equal(t, map[string]string{"AWS_SECRET_ACCESS_KEY": secretKey}, sec.Env)

	// removing a required secret is refused
	in["secrets"] = map[string]string{}
	status, body = a.do(http.MethodPut, path, in)
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "AWS_SECRET_ACCESS_KEY is required")

	// new password and secrets replace the old ones
	in["password"] = "new-password"
	in["secrets"] = map[string]string{"AWS_SECRET_ACCESS_KEY": "rotated-key"}
	status, _ = a.do(http.MethodPut, path, in)
	require.Equal(t, http.StatusOK, status)
	sec = a.secrets(repo.ID)
	assert.Equal(t, "new-password", sec.Password)
	assert.Equal(t, map[string]string{"AWS_SECRET_ACCESS_KEY": "rotated-key"}, sec.Env)

	status, _ = a.do(http.MethodPut, "/repositories/999", s3Input("x"))
	assert.Equal(t, http.StatusNotFound, status)
	a.create("other")
	status, _ = a.do(http.MethodPut, path, s3Input("other"))
	assert.Equal(t, http.StatusConflict, status)
}

func (a *api) secrets(id int64) store.RepositorySecrets {
	a.t.Helper()
	r, err := a.store.GetRepository(context.Background(), id)
	require.NoError(a.t, err)
	box, err := secrets.New(bytes.Repeat([]byte{9}, secrets.KeySize))
	require.NoError(a.t, err)
	sec, err := store.OpenSecrets(box, r.Secrets)
	require.NoError(a.t, err)
	return sec
}

func TestPaths(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID) + "/paths"

	status, body := a.do(http.MethodPut, path, PathsDTO{Include: []string{"relative/dir"}, Exclude: []string{" "}})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), `path \"relative/dir\" is not absolute`)
	assert.Contains(t, string(body), "an exclude pattern is empty")

	status, body = a.do(http.MethodPut, path, PathsDTO{
		Include: []string{"/home/me/Pictures", "/home/me/Documents", "/home/me/Pictures"},
		Exclude: []string{"*.tmp"},
	})
	require.Equal(t, http.StatusOK, status, string(body))
	var dto RepositoryDTO
	a.decode(body, &dto)
	assert.Equal(t, PathsDTO{Include: []string{"/home/me/Documents", "/home/me/Pictures"}, Exclude: []string{"*.tmp"}}, dto.Paths)

	status, body = a.do(http.MethodPut, path, PathsDTO{Include: []string{"/srv"}})
	require.Equal(t, http.StatusOK, status)
	a.decode(body, &dto)
	assert.Equal(t, PathsDTO{Include: []string{"/srv"}, Exclude: []string{}}, dto.Paths, "PUT replaces every path")

	status, _ = a.do(http.MethodPut, "/repositories/999/paths", PathsDTO{})
	assert.Equal(t, http.StatusNotFound, status)
}

func TestRetention(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID) + "/retention"

	status, body := a.do(http.MethodPut, path, RetentionDTO{})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "must keep something")
	status, body = a.do(http.MethodPut, path, RetentionDTO{KeepDaily: -1, KeepWithin: "a while"})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "keep_daily cannot be negative")
	assert.Contains(t, string(body), "keep_within is a duration")

	want := RetentionDTO{KeepDaily: 7, KeepWeekly: 4, KeepYearly: 3, KeepWithin: "1y6m", Prune: true}
	status, body = a.do(http.MethodPut, path, want)
	require.Equal(t, http.StatusOK, status, string(body))
	var dto RepositoryDTO
	a.decode(body, &dto)
	assert.Equal(t, &want, dto.Retention)

	status, body = a.do(http.MethodDelete, path, nil)
	require.Equal(t, http.StatusOK, status)
	a.decode(body, &dto)
	assert.Nil(t, dto.Retention)
}

func TestSchedules(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	base := "/repositories/" + itoa(repo.ID) + "/schedules/"

	status, _ := a.do(http.MethodPut, base+"restore", map[string]any{"cron": "@daily"})
	assert.Equal(t, http.StatusBadRequest, status)
	status, body := a.do(http.MethodPut, base+"backup", map[string]any{"cron": "every day"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid schedule")
	assert.EqualValues(t, 0, a.reloads.n.Load())

	status, body = a.do(http.MethodPut, base+"backup", map[string]any{"cron": "30 2 * * *"})
	require.Equal(t, http.StatusOK, status, string(body))
	status, body = a.do(http.MethodPut, base+"forget", map[string]any{"cron": "@weekly", "enabled": false})
	require.Equal(t, http.StatusOK, status, string(body))
	var dto RepositoryDTO
	a.decode(body, &dto)
	assert.Equal(t, []ScheduleDTO{{Job: "backup", Cron: "30 2 * * *", Enabled: true}, {Job: "forget", Cron: "@weekly", Enabled: false}}, dto.Schedules)
	assert.EqualValues(t, 2, a.reloads.n.Load(), "the scheduler reloads after each change")

	status, body = a.do(http.MethodDelete, base+"backup", nil)
	require.Equal(t, http.StatusOK, status)
	a.decode(body, &dto)
	assert.Equal(t, []ScheduleDTO{{Job: "forget", Cron: "@weekly", Enabled: false}}, dto.Schedules)
	assert.EqualValues(t, 3, a.reloads.n.Load())
}

func TestJobsAndDelete(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID)

	status, _ := a.do(http.MethodPost, path+"/jobs", map[string]any{"kind": "restore"})
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = a.do(http.MethodPost, "/repositories/999/jobs", map[string]any{"kind": "check"})
	assert.Equal(t, http.StatusNotFound, status)

	a.restic.mu.Lock()
	a.restic.block = make(chan struct{})
	a.restic.mu.Unlock()
	status, body := a.do(http.MethodPost, path+"/jobs", map[string]any{"kind": "check"})
	require.Equal(t, http.StatusAccepted, status, string(body))
	var job jobs.JobDTO
	a.decode(body, &job)
	assert.Equal(t, jobs.KindCheck, job.Kind)

	status, body = a.do(http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusConflict, status, "a repository with a running job is kept")
	assert.Contains(t, string(body), "cancel it first")

	close(a.restic.block)
	require.Eventually(t, func() bool {
		_, body := a.do(http.MethodGet, "/jobs/"+itoa(job.ID), nil)
		a.decode(body, &job)
		return job.Status == jobs.StatusSucceeded
	}, 5*time.Second, 10*time.Millisecond)

	status, body = a.do(http.MethodGet, "/jobs?repository="+itoa(repo.ID), nil)
	require.Equal(t, http.StatusOK, status)
	var list []jobs.JobDTO
	a.decode(body, &list)
	assert.Len(t, list, 1)

	status, _ = a.do(http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusNoContent, status)
	assert.EqualValues(t, 1, a.reloads.n.Load())
	status, _ = a.do(http.MethodGet, path, nil)
	assert.Equal(t, http.StatusNotFound, status)
	status, _ = a.do(http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestSnapshots(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID) + "/snapshots"
	a.restic.snapshots = []restic.Snapshot{{ID: "abc", ShortID: "abc"}}

	status, body := a.do(http.MethodGet, path, nil)
	require.Equal(t, http.StatusOK, status, string(body))
	var snaps []restic.Snapshot
	a.decode(body, &snaps)
	assert.Equal(t, a.restic.snapshots, snaps)
	status, _ = a.do(http.MethodGet, path+"?all=true", nil)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []restic.SnapshotFilter{{Host: "nas", Tags: []string{jobs.Tag(repo.ID)}}, {}}, a.restic.filters)

	a.restic.err = &restic.Error{Command: "snapshots", ExitCode: 12}
	status, body = a.do(http.MethodGet, path, nil)
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, string(body), "password does not open")
	a.restic.err = &restic.Error{Command: "snapshots", ExitCode: 10}
	status, body = a.do(http.MethodGet, path, nil)
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, string(body), "not initialized")
}

func TestInvalidIDs(t *testing.T) {
	a := newAPI(t)
	for _, p := range []string{"/repositories/abc", "/repositories/0", "/repositories/-1/snapshots"} {
		status, _ := a.do(http.MethodGet, p, nil)
		assert.Equal(t, http.StatusBadRequest, status, p)
	}
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
