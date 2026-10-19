package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	calls     []string
	nodes     []restic.Node
	found     []restic.FindResult
	dump      string
	groups    []restic.ForgetGroup
	forgets   []restic.ForgetOptions
	exists    bool
	existsErr error
	diff      *restic.Diff
}

func (f *fakeRestic) Exists(_ context.Context, repo *restic.Repo) (bool, error) {
	f.record("exists " + repo.Location)
	return f.exists, f.existsErr
}

func (f *fakeRestic) Diff(_ context.Context, _ *restic.Repo, from, to string) (*restic.Diff, error) {
	f.record("diff " + from + " " + to)
	return f.diff, f.err
}

func (f *fakeRestic) Snapshots(_ context.Context, _ *restic.Repo, filter restic.SnapshotFilter) ([]restic.Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.filters = append(f.filters, filter)
	return f.snapshots, f.err
}

func (f *fakeRestic) record(call string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call)
}

func (f *fakeRestic) Ls(_ context.Context, _ *restic.Repo, snapshotID, dir string) ([]restic.Node, error) {
	f.record("ls " + snapshotID + " " + dir)
	return f.nodes, f.err
}

func (f *fakeRestic) Find(_ context.Context, _ *restic.Repo, snapshotID, pattern string) ([]restic.FindResult, error) {
	f.record("find " + snapshotID + " " + pattern)
	return f.found, f.err
}

func (f *fakeRestic) Stats(_ context.Context, _ *restic.Repo, mode, snapshotID string) (*restic.Stats, error) {
	f.record("stats " + mode + " " + snapshotID)
	return &restic.Stats{TotalSize: 42, SnapshotsCount: 3}, f.err
}

func (f *fakeRestic) Forget(ctx context.Context, _ *restic.Repo, opts restic.ForgetOptions) ([]restic.ForgetGroup, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.forgets = append(f.forgets, opts)
	groups, err := f.groups, f.err
	f.mu.Unlock()
	if werr := f.wait(ctx); werr != nil {
		return nil, werr
	}
	f.mu.Lock()
	return groups, err
}

func (f *fakeRestic) ForgetSnapshots(_ context.Context, _ *restic.Repo, ids []string, prune bool) error {
	f.record(fmt.Sprintf("forget %v prune=%v", ids, prune))
	return f.err
}

func (f *fakeRestic) Dump(_ context.Context, _ *restic.Repo, snapshotID, path, archive string, w io.Writer) error {
	f.record("dump " + snapshotID + " " + path + " " + archive)
	if f.dump != "" {
		if _, err := io.WriteString(w, f.dump); err != nil {
			return err
		}
	}
	return f.err
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
func (f *fakeRestic) Restore(ctx context.Context, _ *restic.Repo, _ restic.RestoreOptions, _ func(restic.RestoreStatus)) (*restic.RestoreSummary, error) {
	return &restic.RestoreSummary{}, f.wait(ctx)
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

func newAPI(t *testing.T, configure ...func(*Server)) *api {
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
	srv := &Server{Store: st, Box: box, Restic: fake, Jobs: manager, Scheduler: a.reloads}
	for _, f := range configure {
		f(srv)
	}
	srv.Register(r.Group(BasePath))
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
	// the local backend has no variables: empty lists, not null
	assert.NotContains(t, string(body), "null")
}

func TestCreateAndGetRepository(t *testing.T) {
	a := newAPI(t)
	created := a.create("photos")
	assert.Equal(t, "photos", created.Name)
	assert.Equal(t, "s3", created.Backend)
	assert.Equal(t, map[string]string{"AWS_DEFAULT_REGION": "eu-west-3", "AWS_ACCESS_KEY_ID": "AKIAEXAMPLE"}, created.Env)
	assert.Equal(t, []string{"AWS_SECRET_ACCESS_KEY"}, created.Secrets)
	assert.True(t, created.HasPassword)
	assert.Nil(t, created.InitializedAt)
	assert.Equal(t, []PlanDTO{}, created.Plans, "plans are added after")
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
		"backend": "s3", "location": "/not/s3",
		"env":      map[string]string{"LD_PRELOAD": "/tmp/x.so"},
		"extended": map[string]string{"--no-lock": "true"},
	})
	require.Equal(t, http.StatusBadRequest, status)
	var e apiError
	a.decode(body, &e)
	details := strings.Join(e.Details, "\n")
	for _, want := range []string{
		"name is required", "password is required", `a s3 location starts with "s3:"`,
		"LD_PRELOAD", "AWS_SECRET_ACCESS_KEY is required", `extended option "--no-lock"`,
	} {
		assert.Contains(t, details, want)
	}

	status, _ = a.do(http.MethodPost, "/repositories", "not an object")
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestOptionsThatRunProgramsNeedCommandHooks(t *testing.T) {
	sftp := map[string]any{
		"name": "sftp", "backend": "sftp", "location": "sftp:nas:/backups", "password": "pw",
		"extended": map[string]string{"sftp.command": "sh -c 'id > /tmp/owned'"},
	}
	a := newAPI(t)
	for _, path := range []string{"/repositories", "/repositories/probe"} {
		status, body := a.do(http.MethodPost, path, sftp)
		require.Equal(t, http.StatusBadRequest, status, path)
		assert.Contains(t, string(body), "sftp.command would run a program", path)
	}
	assert.Nil(t, a.restic.filters, "nothing was run")

	allowed := newAPI(t, func(s *Server) { s.HookCommands = true })
	status, body := allowed.do(http.MethodPost, "/repositories", sftp)
	assert.Equal(t, http.StatusCreated, status, string(body))
}

func TestUpdateRepositoryKeepsOrReplacesSecrets(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID)

	// no password and no secrets: both are kept
	in := s3Input("photos-renamed")
	delete(in, "password")
	delete(in, "secrets")
	status, body := a.do(http.MethodPut, path, in)
	require.Equal(t, http.StatusOK, status, string(body))
	var updated RepositoryDTO
	a.decode(body, &updated)
	assert.Equal(t, "photos-renamed", updated.Name)
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
	tag := jobs.Tag(repo.ID)
	a.restic.snapshots = []restic.Snapshot{
		{ID: "before-plans", Tags: []string{tag}},
		{ID: "of-a-plan", Tags: []string{"work", tag + ":0a1b2c3d"}},
		{ID: "by-hand", Tags: []string{"manual"}},
		{ID: "another-repository", Tags: []string{tag + "0:0a1b2c3d"}},
	}
	require.Nil(t, repo.InitializedAt)

	status, body := a.do(http.MethodGet, path, nil)
	require.Equal(t, http.StatusOK, status, string(body))
	var snaps []restic.Snapshot
	a.decode(body, &snaps)
	assert.Equal(t, a.restic.snapshots[:2], snaps, "the snapshots of every plan, also deleted ones")
	_, body = a.do(http.MethodGet, "/repositories/"+itoa(repo.ID), nil)
	a.decode(body, &repo)
	assert.NotNil(t, repo.InitializedAt, "listing snapshots shows the repository exists")
	status, body = a.do(http.MethodGet, path+"?all=true", nil)
	require.Equal(t, http.StatusOK, status)
	a.decode(body, &snaps)
	assert.Len(t, snaps, 4)
	assert.Equal(t, []restic.SnapshotFilter{{Host: "nas"}, {}}, a.restic.filters)

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
