// Package restic runs the restic binary: one Runner per process, one Repo per
// repository. Credentials only ever travel through the child's environment,
// output is streamed and decoded from restic's --json messages, and every
// call is cancellable through its context.
package restic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Repo is everything needed to reach one restic repository.
type Repo struct {
	// Location is the repository string, e.g. "/srv/restic" or "s3:host/bucket".
	Location string
	Password string
	// Env holds backend variables such as AWS_ACCESS_KEY_ID or B2_ACCOUNT_KEY.
	Env map[string]string
	// CACert is passed as --cacert when set.
	CACert string
	// Options are passed as -o key=value.
	Options map[string]string
}

// Runner executes restic commands.
type Runner struct {
	// Bin is the restic executable, "restic" (looked up in PATH) by default.
	Bin string
	// Env is the base environment of every command; nil means SafeEnv().
	// The repository's variables are added on top of it.
	Env []string
	// RetryLock makes restic wait this long for a repository lock instead of
	// failing at once. Zero disables it.
	RetryLock time.Duration
	// WaitDelay is how long a cancelled command gets to exit after SIGINT
	// before it is killed. Zero means 30s.
	WaitDelay time.Duration
	// StderrLines is how many trailing stderr lines are kept. Zero means 50.
	StderrLines int
}

// NewRunner returns a Runner with the defaults used by ResticGUI.
func NewRunner(bin string) *Runner {
	return &Runner{Bin: bin, RetryLock: 2 * time.Minute}
}

// passedEnv lists the variables of our own environment a restic child may
// see. Anything else (including our own secrets) is dropped.
var passedEnv = map[string]bool{
	"PATH": true, "HOME": true, "USER": true, "TMPDIR": true, "TZ": true, "LANG": true,
	"XDG_CACHE_HOME": true, "RESTIC_CACHE_DIR": true,
	"SSL_CERT_FILE": true, "SSL_CERT_DIR": true,
	"HTTP_PROXY": true, "HTTPS_PROXY": true, "NO_PROXY": true,
	"http_proxy": true, "https_proxy": true, "no_proxy": true,
	"SSH_AUTH_SOCK": true,
	// Windows
	"SYSTEMROOT": true, "SystemRoot": true, "APPDATA": true, "LOCALAPPDATA": true, "USERPROFILE": true, "TEMP": true, "TMP": true,
}

// SafeEnv returns the subset of the current environment that restic needs.
func SafeEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if passedEnv[key] || strings.HasPrefix(key, "LC_") {
			env = append(env, kv)
		}
	}
	return env
}

// Sentinel errors, matched with errors.Is on the *Error a command returns.
var (
	// ErrIncomplete: the backup finished but some files could not be read (exit 3).
	ErrIncomplete = errors.New("restic: some source files could not be read")
	// ErrRepositoryNotFound: there is no repository at the location (exit 10).
	ErrRepositoryNotFound = errors.New("restic: repository does not exist")
	// ErrLocked: the repository is locked by another process (exit 11).
	ErrLocked = errors.New("restic: repository is locked")
	// ErrWrongPassword: the password does not open the repository (exit 12).
	ErrWrongPassword = errors.New("restic: wrong password")
	// ErrInterrupted: restic was interrupted by a signal (exit 130).
	ErrInterrupted = errors.New("restic: interrupted")
)

var exitCodeErrors = map[int]error{
	3:   ErrIncomplete,
	10:  ErrRepositoryNotFound,
	11:  ErrLocked,
	12:  ErrWrongPassword,
	130: ErrInterrupted,
}

// Error is returned when restic fails or is cancelled.
type Error struct {
	Command  string
	ExitCode int // -1 when restic did not exit on its own
	// Message is restic's own error message, when it printed one.
	Message string
	// Stderr holds the last lines restic wrote to stderr.
	Stderr []string
	cause  error
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" && len(e.Stderr) > 0 {
		msg = e.Stderr[len(e.Stderr)-1]
	}
	if msg == "" && e.cause != nil {
		msg = e.cause.Error()
	}
	return fmt.Sprintf("restic %s: exit code %d: %s", e.Command, e.ExitCode, msg)
}

// Unwrap gives access to the context error of a cancelled command.
func (e *Error) Unwrap() error { return e.cause }

// Is matches the sentinel errors by exit code.
func (e *Error) Is(target error) bool {
	return target != nil && exitCodeErrors[e.ExitCode] == target
}

// call describes one restic invocation.
type call struct {
	command string
	args    []string
	// stdout receives restic's standard output when set.
	stdout io.Writer
	// onLine, when set, is called with every stdout line instead.
	onLine func(line []byte)
	// onStderr, when set, is called with every stderr line.
	onStderr func(line string)
}

// run executes c and returns an *Error unless restic exits with code 0.
func (r *Runner) run(ctx context.Context, repo *Repo, c call) error {
	bin := r.Bin
	if bin == "" {
		bin = "restic"
	}
	args := append([]string{c.command}, c.args...)
	if repo != nil {
		args = append(args, repoFlags(repo)...)
		if r.RetryLock > 0 {
			args = append(args, "--retry-lock", r.RetryLock.String())
		}
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = r.environment(repo)
	// SIGINT lets restic release its repository lock before exiting.
	cmd.Cancel = func() error {
		if runtime.GOOS == "windows" {
			return cmd.Process.Kill()
		}
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = r.WaitDelay
	if cmd.WaitDelay == 0 {
		cmd.WaitDelay = 30 * time.Second
	}

	var lines *lineWriter
	switch {
	case c.onLine != nil:
		lines = &lineWriter{onLine: c.onLine}
		cmd.Stdout = lines
	case c.stdout != nil:
		cmd.Stdout = c.stdout
	}
	keep := r.StderrLines
	if keep == 0 {
		keep = 50
	}
	stderr := &stderrTail{max: keep, onLine: c.onStderr}
	cmd.Stderr = stderr

	err := cmd.Run()
	if lines != nil {
		lines.flush()
	}
	if err == nil {
		return nil
	}

	rerr := &Error{Command: c.command, ExitCode: -1, Message: stderr.exitMessage(), Stderr: stderr.lines()}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		rerr.ExitCode = exitErr.ExitCode()
	} else if ctx.Err() == nil {
		// restic could not be started (missing binary, permissions...)
		rerr.cause = err
		rerr.Message = err.Error()
	}
	if ctx.Err() != nil {
		rerr.cause = ctx.Err()
	}
	return rerr
}

func repoFlags(repo *Repo) []string {
	var flags []string
	if repo.CACert != "" {
		flags = append(flags, "--cacert", repo.CACert)
	}
	keys := make([]string, 0, len(repo.Options))
	for k := range repo.Options {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		flags = append(flags, "-o", k+"="+repo.Options[k])
	}
	return flags
}

func (r *Runner) environment(repo *Repo) []string {
	env := r.Env
	if env == nil {
		env = SafeEnv()
	}
	env = append([]string{}, env...)
	if repo == nil {
		return env
	}
	keys := make([]string, 0, len(repo.Env))
	for k := range repo.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		env = append(env, k+"="+repo.Env[k])
	}
	// Last, so a backend variable can never redirect the repository or the
	// password. os/exec keeps the last value of a duplicated key.
	return append(env,
		"RESTIC_REPOSITORY="+repo.Location,
		"RESTIC_PASSWORD="+repo.Password,
		"RESTIC_PASSWORD_FILE=",
		"RESTIC_PASSWORD_COMMAND=",
		"RESTIC_REPOSITORY_FILE=",
	)
}

// lineWriter splits a stream into lines.
type lineWriter struct {
	onLine func([]byte)
	buf    []byte
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			return len(p), nil
		}
		if line := bytes.TrimRight(w.buf[:i], "\r"); len(line) > 0 {
			w.onLine(line)
		}
		w.buf = w.buf[i+1:]
	}
}

func (w *lineWriter) flush() {
	if len(w.buf) > 0 {
		w.onLine(w.buf)
		w.buf = nil
	}
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

// stderrTail keeps the last max lines of stderr, without terminal escapes.
type stderrTail struct {
	mu      sync.Mutex
	max     int
	onLine  func(string)
	partial []byte
	tail    []string
}

func (s *stderrTail) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.partial = append(s.partial, p...)
	for {
		i := bytes.IndexAny(s.partial, "\r\n")
		if i < 0 {
			break
		}
		s.add(string(s.partial[:i]))
		s.partial = s.partial[i+1:]
	}
	return len(p), nil
}

func (s *stderrTail) add(line string) {
	line = strings.TrimSpace(ansiEscape.ReplaceAllString(line, ""))
	if line == "" {
		return
	}
	if s.onLine != nil {
		s.onLine(line)
	}
	s.tail = append(s.tail, line)
	if len(s.tail) > s.max {
		s.tail = s.tail[len(s.tail)-s.max:]
	}
}

func (s *stderrTail) lines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.partial) > 0 {
		s.add(string(s.partial))
		s.partial = nil
	}
	return append([]string{}, s.tail...)
}

// exitMessage returns the message of restic's {"message_type":"exit_error"}
// line, printed on stderr by commands run with --json.
func (s *stderrTail) exitMessage() string {
	lines := s.lines()
	for i := len(lines) - 1; i >= 0; i-- {
		if !strings.HasPrefix(lines[i], "{") {
			continue
		}
		var msg struct {
			Type    string `json:"message_type"`
			Message string `json:"message"`
		}
		if json.Unmarshal([]byte(lines[i]), &msg) == nil && msg.Type == "exit_error" {
			return msg.Message
		}
	}
	return ""
}
