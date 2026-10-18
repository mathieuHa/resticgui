package hooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var run = Run{Event: Success, RepositoryID: 7, Repository: "Photos", JobID: 42, Status: "succeeded", SnapshotID: "f0feba17"}

func TestCommandGetsTheJobInItsEnvironment(t *testing.T) {
	out, err := Runner{Commands: true}.Do(context.Background(), Hook{Kind: Command, Command: `echo "$RESTICGUI_EVENT $RESTICGUI_REPOSITORY $RESTICGUI_REPOSITORY_ID $RESTICGUI_JOB_ID $RESTICGUI_SNAPSHOT_ID"; echo oops >&2`}, run)
	require.NoError(t, err)
	assert.Equal(t, []string{"success Photos 7 42 f0feba17", "oops"}, out)
}

func TestCommandFailure(t *testing.T) {
	out, err := Runner{Commands: true}.Do(context.Background(), Hook{Kind: Command, Command: "echo cannot dump the database; exit 3"}, run)
	assert.EqualError(t, err, "exited with code 3")
	assert.Equal(t, []string{"cannot dump the database"}, out)
}

func TestCommandTimeout(t *testing.T) {
	start := time.Now()
	_, err := Runner{Commands: true}.Do(context.Background(), Hook{Kind: Command, Command: "sleep 30", Timeout: 200 * time.Millisecond}, run)
	assert.EqualError(t, err, "timed out after 200ms")
	assert.Less(t, time.Since(start), 10*time.Second)
}

func TestCommandKeepsTheLastLines(t *testing.T) {
	out, err := Runner{Commands: true}.Do(context.Background(), Hook{Kind: Command, Command: "seq 1 100"}, run)
	require.NoError(t, err)
	require.Len(t, out, MaxOutputLines)
	assert.Equal(t, "100", out[len(out)-1])
}

func TestCommandsOffByDefault(t *testing.T) {
	_, err := Runner{}.Do(context.Background(), Hook{Kind: Command, Command: "true"}, run)
	assert.ErrorIs(t, err, ErrCommandsDisabled)
}

func TestWebhookPostsTheJob(t *testing.T) {
	var got Run
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	out, err := Runner{}.Do(context.Background(), Hook{Kind: Webhook, URL: srv.URL + "/hook?token=secret"}, run)
	require.NoError(t, err)
	assert.Equal(t, run, got)
	assert.Equal(t, []string{"HTTP 204 No Content"}, out)
}

func TestWebhookErrorsDoNotRepeatTheURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	url := srv.URL + "/hook?token=secret"
	_, err := Runner{}.Do(context.Background(), Hook{Kind: Webhook, URL: url}, run)
	assert.EqualError(t, err, "answered 403 Forbidden")

	srv.Close()
	_, err = Runner{}.Do(context.Background(), Hook{Kind: Webhook, URL: url}, run)
	require.Error(t, err)
	assert.False(t, strings.Contains(err.Error(), "secret"), err.Error())
}

func TestCommandDoesNotSeeResticGUISettings(t *testing.T) {
	t.Setenv("RESTICGUI_SECRET_KEY", "the master key")
	t.Setenv("HOOK_TEST_VAR", "kept")
	out, err := Runner{Commands: true}.Do(context.Background(), Hook{Kind: Command, Command: `env | grep -E '^(RESTICGUI_|HOOK_TEST_)' | sort`}, run)
	require.NoError(t, err)
	joined := strings.Join(out, "\n")
	assert.NotContains(t, joined, "the master key")
	assert.Contains(t, joined, "HOOK_TEST_VAR=kept")
	assert.Contains(t, joined, "RESTICGUI_JOB_ID=42", "the job variables are still there")
}

func TestCommandTimeoutEndsItsChildren(t *testing.T) {
	marker := t.TempDir() + "/still-running"
	start := time.Now()
	_, err := Runner{Commands: true}.Do(context.Background(), Hook{Kind: Command,
		Command: "(sleep 1; touch " + marker + ") & sleep 30", Timeout: 200 * time.Millisecond}, run)
	assert.EqualError(t, err, "timed out after 200ms")
	assert.Less(t, time.Since(start), 10*time.Second)
	time.Sleep(1500 * time.Millisecond)
	assert.NoFileExists(t, marker, "the background child was killed with the shell")
}

func TestCommandOutputIsBounded(t *testing.T) {
	// 4 MB on one line, then a last line
	out, err := Runner{Commands: true}.Do(context.Background(), Hook{Kind: Command,
		Command: "head -c 4194304 /dev/zero | tr '\\0' x; echo; echo done"}, run)
	require.NoError(t, err)
	assert.Equal(t, "done", out[len(out)-1])
	total := 0
	for _, l := range out {
		total += len(l)
	}
	assert.LessOrEqual(t, total, maxOutputBytes)
}

func TestTailKeepsTheEnd(t *testing.T) {
	tl := &tail{max: 8}
	for _, s := range []string{"abc", "defgh", "ijk", "0123456789"} {
		_, _ = tl.Write([]byte(s))
	}
	assert.Equal(t, "23456789", string(tl.b))
	tl = &tail{max: 8}
	_, _ = tl.Write([]byte("abcdef"))
	_, _ = tl.Write([]byte("ghij"))
	assert.Equal(t, "cdefghij", string(tl.b))
}

func TestWebhookDoesNotFollowRedirects(t *testing.T) {
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the redirect was followed")
	}))
	defer internal.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.URL+"/latest/meta-data/", http.StatusFound)
	}))
	defer srv.Close()
	out, err := Runner{}.Do(context.Background(), Hook{Kind: Webhook, URL: srv.URL}, run)
	assert.EqualError(t, err, "answered 302 Found")
	assert.Equal(t, []string{"HTTP 302 Found"}, out)
}
