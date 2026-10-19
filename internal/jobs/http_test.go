package jobs

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newServer(t *testing.T, e *env) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&Handler{Manager: e.manager, Heartbeat: 20 * time.Millisecond}).Register(r)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, method, url string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, body
}

func TestHTTPBackupCancelAndList(t *testing.T) {
	e := newEnv(t, Options{})
	g := newGate()
	e.restic.run = g.run
	srv := newServer(t, e)
	repo := e.repository(t, "r")
	e.addPath(t, repo.ID, "/data", "include")

	resp, body := do(t, http.MethodPost, srv.URL+"/repositories/"+strconv.FormatInt(repo.ID, 10)+"/backup")
	require.Equal(t, http.StatusAccepted, resp.StatusCode, string(body))
	var job JobDTO
	require.NoError(t, json.Unmarshal(body, &job))
	assert.Equal(t, KindBackup, job.Kind)
	assert.Equal(t, TriggerManual, job.Trigger)
	assert.Equal(t, StatusQueued, job.Status)

	g.waitStarted(t)
	jobURL := srv.URL + "/jobs/" + strconv.FormatInt(job.ID, 10)
	resp, _ = do(t, http.MethodPost, jobURL+"/cancel")
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	e.waitStatus(t, job.ID, StatusCancelled)
	resp, _ = do(t, http.MethodPost, jobURL+"/cancel")
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	resp, body = do(t, http.MethodGet, jobURL)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.Unmarshal(body, &job))
	assert.Equal(t, StatusCancelled, job.Status)
	assert.JSONEq(t, `{}`, string(job.Summary))

	resp, body = do(t, http.MethodGet, srv.URL+"/jobs?repository="+strconv.FormatInt(repo.ID, 10))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list []JobDTO
	require.NoError(t, json.Unmarshal(body, &list))
	require.Len(t, list, 1)
	assert.Equal(t, job.ID, list[0].ID)
}

func TestHTTPErrors(t *testing.T) {
	e := newEnv(t, Options{})
	srv := newServer(t, e)
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{http.MethodPost, "/repositories/999/backup", http.StatusNotFound},
		{http.MethodPost, "/repositories/abc/backup", http.StatusBadRequest},
		{http.MethodPost, "/jobs/999/cancel", http.StatusNotFound},
		{http.MethodGet, "/jobs/999", http.StatusNotFound},
		{http.MethodGet, "/jobs/0", http.StatusBadRequest},
		{http.MethodGet, "/jobs", http.StatusBadRequest},
		{http.MethodGet, "/jobs?repository=1&limit=0", http.StatusBadRequest},
		{http.MethodGet, "/jobs?repository=1&offset=-1", http.StatusBadRequest},
	} {
		resp, body := do(t, tc.method, srv.URL+tc.path)
		assert.Equal(t, tc.status, resp.StatusCode, "%s %s: %s", tc.method, tc.path, body)
	}
}

func TestHTTPEventStream(t *testing.T) {
	e := newEnv(t, Options{})
	srv := newServer(t, e)
	repo := e.repository(t, "r")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/events", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	lines := make(chan string, 64)
	go func() {
		defer close(lines)
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			lines <- sc.Text()
		}
	}()
	next := func() string {
		t.Helper()
		select {
		case l, ok := <-lines:
			require.True(t, ok, "stream ended")
			return l
		case <-time.After(5 * time.Second):
			t.Fatal("no line on the stream")
			return ""
		}
	}
	assert.Equal(t, "retry: 3000", next())
	assert.Equal(t, "", next())
	assert.Equal(t, ": ping", next(), "heartbeat while idle")
	assert.Equal(t, "", next())

	job, err := e.manager.Enqueue(context.Background(), repo.ID, 0, KindCheck, TriggerManual)
	require.NoError(t, err)
	var statuses []Status
	for len(statuses) < 3 {
		l := next()
		if !strings.HasPrefix(l, "event: ") {
			continue
		}
		assert.Equal(t, "event: job", l)
		data := strings.TrimPrefix(next(), "data: ")
		var ev Event
		require.NoError(t, json.Unmarshal([]byte(data), &ev), data)
		assert.Equal(t, job.ID, ev.JobID)
		assert.Equal(t, repo.ID, ev.RepositoryID)
		assert.Equal(t, KindCheck, ev.Kind)
		statuses = append(statuses, ev.Status)
	}
	assert.Equal(t, []Status{StatusQueued, StatusRunning, StatusSucceeded}, statuses)
}

func TestHTTPEventStreamOutlivesWriteTimeout(t *testing.T) {
	e := newEnv(t, Options{})
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&Handler{Manager: e.manager, Heartbeat: 150 * time.Millisecond}).Register(r)
	srv := httptest.NewUnstartedServer(r)
	srv.Config.WriteTimeout = 50 * time.Millisecond
	srv.Start()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/events")
	require.NoError(t, err)
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	var lines []string
	for len(lines) < 4 && sc.Scan() {
		lines = append(lines, sc.Text())
	}
	assert.Equal(t, []string{"retry: 3000", "", ": ping", ""}, lines, "the heartbeat arrives after the write timeout")
}
