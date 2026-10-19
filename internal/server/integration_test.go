//go:build integration

package server

// The API against a real restic binary and a local repository:
//
//	RESTIC_BIN=/path/to/restic go test -tags integration ./internal/server/

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestIntegrationAPI(t *testing.T) {
	bin := os.Getenv("RESTIC_BIN")
	if bin == "" {
		bin = "restic"
	}
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "v2.db"))
	require.NoError(t, err)
	defer st.Close()
	box, err := secrets.New(bytes.Repeat([]byte{1}, secrets.KeySize))
	require.NoError(t, err)
	runner := restic.NewRunner(bin)
	manager := jobs.New(st, box, runner, jobs.NewBroker(), jobs.Options{})
	defer manager.Close()
	r := gin.New()
	(&Server{Store: st, Box: box, Restic: runner, Jobs: manager, Scheduler: &reloads{}, Roots: Roots{root}}).Register(r.Group(BasePath))
	srv := httptest.NewServer(r)
	defer srv.Close()
	a := &api{t: t, srv: srv, store: st}

	src := filepath.Join(root, "src")
	require.NoError(t, os.MkdirAll(filepath.Join(src, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "docs", "report.txt"), []byte("quarterly report"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "photo.jpg"), []byte("not a jpeg"), 0o644))

	status, body := a.do(http.MethodPost, "/repositories", map[string]any{
		"name": "local", "backend": "local", "location": filepath.Join(root, "repo"), "password": password, "host": "it",
	})
	require.Equal(t, http.StatusCreated, status, string(body))
	var repo RepositoryDTO
	a.decode(body, &repo)
	base := "/repositories/" + itoa(repo.ID)
	status, body = a.do(http.MethodPost, base+"/plans", planInput("Everything", src))
	require.Equal(t, http.StatusCreated, status, string(body))

	run := func(kind string) jobs.JobDTO {
		t.Helper()
		status, body := a.do(http.MethodPost, base+"/jobs", map[string]any{"kind": kind})
		require.Equal(t, http.StatusAccepted, status, string(body))
		var job jobs.JobDTO
		a.decode(body, &job)
		return a.waitJob(job.ID)
	}
	run("init")
	backup := run("backup")
	require.NotEmpty(t, backup.SnapshotID)
	snap := backup.SnapshotID[:8]

	status, body = a.do(http.MethodGet, base+"/snapshots/"+snap+"/files?path="+src, nil)
	require.Equal(t, http.StatusOK, status, string(body))
	var nodes []restic.Node
	a.decode(body, &nodes)
	var names []string
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	assert.ElementsMatch(t, []string{"docs", "photo.jpg"}, names)

	status, body = a.do(http.MethodGet, base+"/snapshots/latest/find?pattern=*.txt", nil)
	require.Equal(t, http.StatusOK, status, string(body))
	assert.Contains(t, string(body), "report.txt")

	resp, err := http.Get(srv.URL + BasePath + base + "/snapshots/" + snap + "/download?path=" + filepath.Join(src, "docs"))
	require.NoError(t, err)
	archive, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(archive))
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)
	var zipped []string
	for _, f := range zr.File {
		zipped = append(zipped, f.Name)
	}
	// restic keeps the full path of each file inside the archive
	require.Len(t, zipped, 1)
	assert.Equal(t, "docs/report.txt", filepath.ToSlash(zipped[0])[len(zipped[0])-len("docs/report.txt"):])

	target := filepath.Join(root, "restored")
	status, body = a.do(http.MethodPost, base+"/snapshots/"+snap+"/restore", map[string]any{"target": target})
	require.Equal(t, http.StatusAccepted, status, string(body))
	var job jobs.JobDTO
	a.decode(body, &job)
	a.waitJob(job.ID)
	restored, err := os.ReadFile(filepath.Join(target, src, "docs", "report.txt"))
	require.NoError(t, err)
	assert.Equal(t, "quarterly report", string(restored))

	status, body = a.do(http.MethodGet, base+"/stats", nil)
	require.Equal(t, http.StatusOK, status, string(body))
	assert.Contains(t, string(body), `"snapshots_count":1`)

	status, body = a.do(http.MethodDelete, base+"/snapshots/"+snap+"?prune=true", nil)
	require.Equal(t, http.StatusNoContent, status, string(body))
	status, body = a.do(http.MethodGet, base+"/snapshots", nil)
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `[]`, string(body))
}

func (a *api) waitJob(id int64) jobs.JobDTO {
	a.t.Helper()
	var job jobs.JobDTO
	require.Eventually(a.t, func() bool {
		_, body := a.do(http.MethodGet, "/jobs/"+itoa(id), nil)
		a.decode(body, &job)
		return job.FinishedAt != nil
	}, 2*time.Minute, 50*time.Millisecond)
	require.Equal(a.t, jobs.StatusSucceeded, job.Status, "%s: %s\n%s", job.Kind, job.Error, job.LogExcerpt)
	return job
}
