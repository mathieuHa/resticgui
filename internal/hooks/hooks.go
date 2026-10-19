// Package hooks runs the commands and webhooks set on a repository around
// its backups: before (a failure stops the backup), and after a success or
// a failure.
package hooks

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Events, as stored in hooks.event.
const (
	Before  = "before"
	Success = "success"
	Failure = "failure"
)

// Kinds, as stored in hooks.kind.
const (
	Command = "command"
	Webhook = "webhook"
)

// ErrCommandsDisabled: a command hook exists but the configuration does not
// allow commands (hooks.commands).
var ErrCommandsDisabled = errors.New("hooks: commands are turned off (hooks.commands in the configuration)")

// MaxOutputLines is how many of the last lines of a hook's output are kept
// in the job log.
const MaxOutputLines = 20

// Hook is one command or webhook.
type Hook struct {
	Event   string
	Kind    string
	Command string
	URL     string
	Timeout time.Duration
}

// Run describes the job a hook runs for. Commands get it as environment
// variables (RESTICGUI_*), webhooks as a JSON body.
type Run struct {
	Event        string `json:"event"`
	RepositoryID int64  `json:"repository_id"`
	Repository   string `json:"repository"`
	// Plan is the name of the backup plan that ran.
	Plan  string `json:"plan"`
	JobID int64  `json:"job_id"`
	// Status is the job's status after it ran: succeeded or failed.
	Status     string `json:"status,omitempty"`
	SnapshotID string `json:"snapshot_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Runner runs hooks.
type Runner struct {
	// Commands allows command hooks; without it they fail with ErrCommandsDisabled.
	Commands bool
	// Shell runs commands as Shell -c COMMAND; default /bin/sh.
	Shell string
	// HTTP sends webhooks; default http.DefaultClient (the hook's timeout applies).
	HTTP *http.Client
}

// Do runs h and returns the last lines of its output (a webhook's response
// status) and an error when it fails, times out or exits non-zero.
func (r Runner) Do(ctx context.Context, h Hook, run Run) ([]string, error) {
	timeout := h.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var (
		out []string
		err error
	)
	switch h.Kind {
	case Command:
		out, err = r.command(ctx, h.Command, run)
	case Webhook:
		out, err = r.webhook(ctx, h.URL, run)
	default:
		return nil, fmt.Errorf("hooks: unknown kind %q", h.Kind)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return out, fmt.Errorf("timed out after %s", timeout)
	}
	return out, err
}

func (r Runner) command(ctx context.Context, command string, run Run) ([]string, error) {
	if !r.Commands {
		return nil, ErrCommandsDisabled
	}
	shell := r.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	// #nosec G204 -- running the configured command is the point of a command
	// hook; they are off unless hooks.commands is set
	cmd := exec.CommandContext(ctx, shell, "-c", command)
	cmd.Env = append(hookEnv(),
		"RESTICGUI_EVENT="+run.Event,
		"RESTICGUI_REPOSITORY_ID="+strconv.FormatInt(run.RepositoryID, 10),
		"RESTICGUI_REPOSITORY="+run.Repository,
		"RESTICGUI_PLAN="+run.Plan,
		"RESTICGUI_JOB_ID="+strconv.FormatInt(run.JobID, 10),
		"RESTICGUI_STATUS="+run.Status,
		"RESTICGUI_SNAPSHOT_ID="+run.SnapshotID,
		"RESTICGUI_ERROR="+run.Error,
	)
	// a command that leaves children holding its output open still ends
	cmd.WaitDelay = 2 * time.Second
	// a timeout ends the command's children too
	killGroup(cmd)
	buf := &tail{max: maxOutputBytes}
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	out := lastLines(bytes.NewReader(buf.b), MaxOutputLines)
	var exit *exec.ExitError
	if errors.As(err, &exit) && ctx.Err() == nil {
		return out, fmt.Errorf("exited with code %d", exit.ExitCode())
	}
	return out, err
}

func (r Runner) webhook(ctx context.Context, url string, run Run) ([]string, error) {
	body, err := json.Marshal(run)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("invalid webhook URL")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ResticGUI")
	client := r.HTTP
	if client == nil {
		client = noRedirects
	}
	res, err := client.Do(req)
	if err != nil {
		// the error repeats the URL, which can carry a token
		var uerr interface{ Unwrap() error }
		if errors.As(err, &uerr) && uerr.Unwrap() != nil {
			return nil, fmt.Errorf("request failed: %w", uerr.Unwrap())
		}
		return nil, errors.New("request failed")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
	out := []string{"HTTP " + res.Status}
	if res.StatusCode >= 300 {
		return out, fmt.Errorf("answered %s", res.Status)
	}
	return out, nil
}

// noRedirects posts webhooks without following redirects, so that a hook
// cannot be bounced to an address of the server's network.
var noRedirects = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

// hookEnv is ResticGUI's environment without its own settings, among which
// the key that decrypts every stored password.
func hookEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "RESTICGUI_") {
			env = append(env, kv)
		}
	}
	return env
}

// maxOutputBytes is how much of a command's output is kept, its end.
const maxOutputBytes = 256 << 10

// tail keeps the last max bytes written to it.
type tail struct {
	max int
	b   []byte
}

func (t *tail) Write(p []byte) (int, error) {
	n := len(p)
	if len(p) >= t.max {
		t.b = append(t.b[:0], p[len(p)-t.max:]...)
		return n, nil
	}
	if over := len(t.b) + len(p) - t.max; over > 0 {
		t.b = append(t.b[:0], t.b[over:]...)
	}
	t.b = append(t.b, p...)
	return n, nil
}

func lastLines(r io.Reader, n int) []string {
	var lines []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		lines = append(lines, strings.TrimRight(sc.Text(), "\r"))
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	return lines
}
