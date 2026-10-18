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
