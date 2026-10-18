package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/notify"
	"github.com/mathieuHa/resticgui/internal/restic"
)

// a notification URL whose token is a value do() checks never leaks
var ntfyURL = "ntfy://" + secretKey + "@ntfy.sh/backups"

type sent struct {
	mu   sync.Mutex
	msgs []string
	err  error
}

func withNotify(s *sent) func(*Server) {
	return func(srv *Server) {
		srv.Notify = &notify.Service{Store: srv.Store, Box: srv.Box, Events: jobs.NewBroker(), Send: func(rawURL, title, body string) error {
			s.mu.Lock()
			defer s.mu.Unlock()
			s.msgs = append(s.msgs, rawURL+" | "+title)
			if s.err != nil {
				return notify.Redact(s.err, rawURL)
			}
			return nil
		}}
	}
}

func TestSettings(t *testing.T) {
	a := newAPI(t, func(s *Server) { s.HookCommands = true })
	status, body := a.do(http.MethodGet, "/settings", nil)
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `{"hook_commands":true,"metrics":false}`, string(body))
}

func TestNotifiers(t *testing.T) {
	var s sent
	a := newAPI(t, withNotify(&s))
	status, body := a.do(http.MethodPost, "/notifiers", map[string]any{"name": "Phone", "url": ntfyURL, "on_failure": true, "on_missed": true, "enabled": true})
	require.Equal(t, http.StatusCreated, status, string(body))
	var n NotifierDTO
	a.decode(body, &n)
	assert.Equal(t, "ntfy", n.Service)
	assert.True(t, n.OnFailure)
	assert.False(t, n.OnSuccess)

	status, body = a.do(http.MethodPost, "/notifiers", map[string]any{"name": "Phone", "url": ntfyURL})
	assert.Equal(t, http.StatusConflict, status, string(body))
	status, body = a.do(http.MethodPost, "/notifiers", map[string]any{"name": "Other", "url": "nosuch://" + secretKey + "@host"})
	assert.Equal(t, http.StatusBadRequest, status, string(body))
	status, _ = a.do(http.MethodPost, "/notifiers", map[string]any{"name": "No URL"})
	assert.Equal(t, http.StatusBadRequest, status)

	// an update without a URL keeps it
	status, body = a.do(http.MethodPut, "/notifiers/"+itoa(n.ID), map[string]any{"name": "My phone", "on_success": true, "enabled": false})
	require.Equal(t, http.StatusOK, status, string(body))
	a.decode(body, &n)
	assert.Equal(t, "My phone", n.Name)
	assert.Equal(t, "ntfy", n.Service)
	assert.False(t, n.Enabled)

	status, _ = a.do(http.MethodPost, "/notifiers/"+itoa(n.ID)+"/test", nil)
	require.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, []string{ntfyURL + " | 🔔 ResticGUI test"}, s.msgs, "the stored URL is used")

	s.err = errors.New("POST https://ntfy.sh/backups with token " + secretKey + ": 401")
	status, body = a.do(http.MethodPost, "/notifiers/"+itoa(n.ID)+"/test", nil)
	assert.Equal(t, http.StatusBadGateway, status)
	assert.Contains(t, string(body), "could not send")

	status, body = a.do(http.MethodGet, "/notifiers", nil)
	require.Equal(t, http.StatusOK, status)
	assert.NotContains(t, string(body), "ntfy.sh")

	status, _ = a.do(http.MethodDelete, "/notifiers/"+itoa(n.ID), nil)
	assert.Equal(t, http.StatusNoContent, status)
	status, _ = a.do(http.MethodDelete, "/notifiers/"+itoa(n.ID), nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestHealthcheck(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	assert.Nil(t, repo.Healthcheck)
	assert.Equal(t, []HookDTO{}, repo.Hooks)

	status, body := a.do(http.MethodPut, "/repositories/"+itoa(repo.ID)+"/healthcheck", map[string]any{"url": "https://hc-ping.com/" + secretKey})
	require.Equal(t, http.StatusOK, status, string(body))
	a.decode(body, &repo)
	assert.Equal(t, &HealthcheckDTO{Host: "hc-ping.com"}, repo.Healthcheck)

	status, _ = a.do(http.MethodPut, "/repositories/"+itoa(repo.ID)+"/healthcheck", map[string]any{"url": "ftp://example.com"})
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = a.do(http.MethodPut, "/repositories/999/healthcheck", map[string]any{"url": "https://hc-ping.com/x"})
	assert.Equal(t, http.StatusNotFound, status)

	status, body = a.do(http.MethodDelete, "/repositories/"+itoa(repo.ID)+"/healthcheck", nil)
	require.Equal(t, http.StatusOK, status)
	a.decode(body, &repo)
	assert.Nil(t, repo.Healthcheck)
}

func TestHooks(t *testing.T) {
	a := newAPI(t, func(s *Server) { s.HookCommands = true })
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID) + "/hooks"
	status, body := a.do(http.MethodPut, path, []map[string]any{
		{"event": "success", "kind": "webhook", "url": "https://hooks.example.com/done?token=" + secretKey},
		{"event": "before", "kind": "command", "command": "  pg_dump app > /backup/app.sql  ", "timeout_seconds": 600},
	})
	require.Equal(t, http.StatusOK, status, string(body))
	a.decode(body, &repo)
	require.Len(t, repo.Hooks, 2)
	assert.Equal(t, HookDTO{ID: repo.Hooks[0].ID, Event: "before", Kind: "command", Command: "pg_dump app > /backup/app.sql", TimeoutSeconds: 600}, repo.Hooks[0], "before hooks first")
	assert.Equal(t, HookDTO{ID: repo.Hooks[1].ID, Event: "success", Kind: "webhook", Host: "hooks.example.com", TimeoutSeconds: 300}, repo.Hooks[1])

	// the webhook keeps its URL when sent back without one
	webhook := repo.Hooks[1].ID
	status, body = a.do(http.MethodPut, path, []map[string]any{{"id": webhook, "event": "failure", "kind": "webhook"}})
	require.Equal(t, http.StatusOK, status, string(body))
	a.decode(body, &repo)
	require.Len(t, repo.Hooks, 1)
	assert.Equal(t, "hooks.example.com", repo.Hooks[0].Host)
	stored, err := a.store.ListHooks(context.Background(), repo.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, stored[0].Url)

	status, body = a.do(http.MethodPut, path, []map[string]any{
		{"event": "after", "kind": "command", "command": "true"},
		{"event": "before", "kind": "webhook"},
		{"event": "before", "kind": "command", "command": " "},
		{"event": "before", "kind": "script", "timeout_seconds": -1},
	})
	require.Equal(t, http.StatusBadRequest, status)
	var e apiError
	a.decode(body, &e)
	assert.Equal(t, []string{
		"hook 1: event is one of before, success, failure",
		"hook 2: url is required",
		"hook 3: command is required",
		"hook 4: timeout_seconds is between 1 and 86400",
		"hook 4: kind is command or webhook",
	}, e.Details)

	status, _ = a.do(http.MethodPut, path, []map[string]any{})
	require.Equal(t, http.StatusOK, status)
}

func TestCommandHooksNeedTheConfiguration(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	status, body := a.do(http.MethodPut, "/repositories/"+itoa(repo.ID)+"/hooks", []map[string]any{{"event": "before", "kind": "command", "command": "true"}})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "hooks.commands")
}

func TestProbe(t *testing.T) {
	a := newAPI(t)
	a.restic.exists = true
	a.restic.snapshots = []restic.Snapshot{
		{ID: "a1", Hostname: "old", Paths: []string{"/old"}},
		{ID: "b2", Hostname: "nas", Paths: []string{"/home/me/Pictures"}},
	}
	status, body := a.do(http.MethodPost, "/repositories/probe", s3Input("photos"))
	require.Equal(t, http.StatusOK, status, string(body))
	var p ProbeDTO
	a.decode(body, &p)
	assert.True(t, p.Exists)
	assert.Equal(t, 2, p.Snapshots)
	assert.Equal(t, "nas", p.Latest.Hostname)
	assert.Equal(t, []restic.SnapshotFilter{{}}, a.restic.filters, "every snapshot, not only resticgui's")
	repos, err := a.store.ListRepositories(context.Background())
	require.NoError(t, err)
	assert.Empty(t, repos, "a probe saves nothing")

	a.restic.exists = false
	status, body = a.do(http.MethodPost, "/repositories/probe", s3Input("photos"))
	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `{"exists":false,"snapshots":0,"latest":null}`, string(body))

	a.restic.existsErr = restic.ErrWrongPassword
	status, body = a.do(http.MethodPost, "/repositories/probe", s3Input("photos"))
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, string(body), "password does not open it")

	in := s3Input("photos")
	delete(in, "password")
	status, _ = a.do(http.MethodPost, "/repositories/probe", in)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestDiff(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	a.restic.diff = &restic.Diff{Changes: []restic.DiffChange{{Path: "/a", Modifier: "+"}}, Stats: restic.DiffStats{ChangedFiles: 0, Added: restic.DiffCount{Files: 1}}}
	status, body := a.do(http.MethodGet, "/repositories/"+itoa(repo.ID)+"/snapshots/1135e1d8/diff?to=f7108643", nil)
	require.Equal(t, http.StatusOK, status, string(body))
	var d restic.Diff
	a.decode(body, &d)
	assert.Equal(t, *a.restic.diff, d)
	assert.Contains(t, a.restic.calls, "diff 1135e1d8 f7108643")

	status, _ = a.do(http.MethodGet, "/repositories/"+itoa(repo.ID)+"/snapshots/1135e1d8/diff?to=--json", nil)
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = a.do(http.MethodGet, "/repositories/"+itoa(repo.ID)+"/snapshots/1135e1d8/diff", nil)
	assert.Equal(t, http.StatusBadRequest, status)
}
