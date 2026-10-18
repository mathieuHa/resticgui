package restic

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironment(t *testing.T) {
	t.Setenv("RESTICGUI_SECRET_KEY", "must-not-leak")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "from-our-own-environment")
	t.Setenv("LC_ALL", "C.UTF-8")

	f := newFake(t, scenario{})
	f.runner.Env = nil // use SafeEnv
	f.repo.Env["AWS_ACCESS_KEY_ID"] = "AKIAEXAMPLE"
	f.repo.Env["RESTIC_PASSWORD"] = "attempt-to-override"
	f.repo.Env["RESTIC_REPOSITORY"] = "/somewhere/else"
	require.NoError(t, f.runner.Unlock(context.Background(), f.repo))

	env := f.invocation().Env
	assert.Equal(t, "repository-password", env["RESTIC_PASSWORD"])
	assert.Equal(t, "/srv/restic/docs", env["RESTIC_REPOSITORY"])
	assert.Equal(t, "AKIAEXAMPLE", env["AWS_ACCESS_KEY_ID"])
	assert.Equal(t, "C.UTF-8", env["LC_ALL"])
	assert.NotEmpty(t, env["PATH"])
	assert.NotContains(t, env, "RESTICGUI_SECRET_KEY")
	assert.NotContains(t, env, "AWS_SECRET_ACCESS_KEY", "only the repository's own credentials are passed")
	assert.Empty(t, env["RESTIC_PASSWORD_FILE"])
	assert.Empty(t, env["RESTIC_PASSWORD_COMMAND"])
}

func TestArgumentsNeverCarrySecrets(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/backup.jsonl"})
	f.repo.Env["AWS_SECRET_ACCESS_KEY"] = "s3-secret"
	f.repo.CACert = "/etc/ssl/minio.pem"
	f.repo.Options = map[string]string{"s3.storage-class": "STANDARD_IA", "s3.region": "eu-west-3"}
	f.runner.RetryLock = 2 * time.Minute

	_, err := f.runner.Backup(context.Background(), f.repo, BackupOptions{Paths: []string{"/home/me"}}, nil)
	require.NoError(t, err)

	args := f.invocation().Args
	for _, a := range args {
		assert.NotContains(t, a, "repository-password")
		assert.NotContains(t, a, "s3-secret")
		assert.NotContains(t, a, "/srv/restic/docs", "the location goes through RESTIC_REPOSITORY")
	}
	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "--cacert /etc/ssl/minio.pem")
	assert.Contains(t, joined, "-o s3.region=eu-west-3 -o s3.storage-class=STANDARD_IA")
	assert.Contains(t, joined, "--retry-lock 2m0s")
	assert.Equal(t, "backup", args[0])
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		exit    int
		stderr  []string
		target  error
		message string
	}{
		{10, []string{`{"message_type":"exit_error","code":10,"message":"Fatal: repository does not exist: unable to open config file"}`}, ErrRepositoryNotFound, "repository does not exist"},
		{11, []string{"some text", `{"message_type":"exit_error","code":11,"message":"unable to create lock in backend: repository is already locked by PID 13687"}`}, ErrLocked, "already locked"},
		{12, []string{`{"message_type":"exit_error","code":12,"message":"Fatal: wrong password or no key found"}`}, ErrWrongPassword, "wrong password"},
		{130, nil, ErrInterrupted, ""},
		{1, []string{"plain text failure"}, nil, "plain text failure"},
	}
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			f := newFake(t, scenario{Exit: tt.exit, Stderr: tt.stderr})
			err := f.runner.Unlock(context.Background(), f.repo)
			var rerr *Error
			require.ErrorAs(t, err, &rerr)
			assert.Equal(t, tt.exit, rerr.ExitCode)
			assert.Equal(t, "unlock", rerr.Command)
			if tt.target != nil {
				assert.ErrorIs(t, err, tt.target)
			}
			for _, other := range []error{ErrRepositoryNotFound, ErrLocked, ErrWrongPassword, ErrIncomplete} {
				if other != tt.target {
					assert.NotErrorIs(t, err, other)
				}
			}
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestCancelSendsSIGINT(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/backup-progress.jsonl", Block: true, SigintExit: 1})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	progressed := make(chan struct{}, 10)
	done := make(chan error, 1)
	go func() {
		_, err := f.runner.Backup(ctx, f.repo, BackupOptions{Paths: []string{"/home/me"}}, func(BackupStatus) {
			progressed <- struct{}{}
		})
		done <- err
	}()

	select {
	case <-progressed:
	case <-time.After(10 * time.Second):
		t.Fatal("no progress received")
	}
	cancel()

	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.Canceled)
		var rerr *Error
		require.ErrorAs(t, err, &rerr)
		assert.Equal(t, 1, rerr.ExitCode)
		assert.Equal(t, []string{"signal interrupt received, cleaning up"}, rerr.Stderr, "terminal escapes are stripped")
	case <-time.After(10 * time.Second):
		t.Fatal("cancelled backup did not return")
	}
	assert.True(t, f.invocation().GotSigint)
}

func TestCancelKillsAfterWaitDelay(t *testing.T) {
	f := newFake(t, scenario{Block: true, IgnoreSigint: true})
	f.runner.WaitDelay = 200 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := f.runner.Prune(ctx, f.repo)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 5*time.Second)
	assert.True(t, f.invocation().GotSigint, "SIGINT is tried before killing")
}

func TestMissingBinary(t *testing.T) {
	r := &Runner{Bin: "/nonexistent/restic"}
	err := r.Unlock(context.Background(), &Repo{Location: "/x", Password: "p"})
	var rerr *Error
	require.ErrorAs(t, err, &rerr)
	assert.Equal(t, -1, rerr.ExitCode)
	assert.Contains(t, err.Error(), "no such file")
	assert.False(t, errors.Is(err, context.Canceled))
}

func TestStderrIsBounded(t *testing.T) {
	var lines []string
	for i := 0; i < 200; i++ {
		lines = append(lines, strings.Repeat("x", 10))
	}
	f := newFake(t, scenario{Exit: 1, Stderr: append(lines, "last line")})
	f.runner.StderrLines = 5
	err := f.runner.Prune(context.Background(), f.repo)
	var rerr *Error
	require.ErrorAs(t, err, &rerr)
	assert.Len(t, rerr.Stderr, 5)
	assert.Equal(t, "last line", rerr.Stderr[4])
}

func TestLineWriterSplitsAcrossWrites(t *testing.T) {
	var got []string
	w := &lineWriter{onLine: func(l []byte) { got = append(got, string(l)) }}
	for _, chunk := range []string{`{"a":`, "1}\n{\"b\"", ":2}\r\n\n", `{"c":3}`} {
		_, err := w.Write([]byte(chunk))
		require.NoError(t, err)
	}
	w.flush()
	assert.Equal(t, []string{`{"a":1}`, `{"b":2}`, `{"c":3}`}, got)
}

func TestSafeEnv(t *testing.T) {
	t.Setenv("HOME", "/home/me")
	t.Setenv("MY_TOKEN", "secret")
	env := strings.Join(SafeEnv(), "\n")
	assert.Contains(t, env, "HOME=/home/me")
	assert.NotContains(t, env, "MY_TOKEN")
}

func TestCommandOptionsNeedPermission(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/backup.jsonl"})
	f.repo.Options = map[string]string{"sftp.command": "sh -c 'touch /tmp/owned'"}
	_, err := f.runner.Backup(context.Background(), f.repo, BackupOptions{Paths: []string{"/home/me"}}, nil)
	require.ErrorIs(t, err, ErrCommandsOff)
	assert.NoFileExists(t, f.record, "restic was not started")

	f.runner.AllowCommands = true
	_, err = f.runner.Backup(context.Background(), f.repo, BackupOptions{Paths: []string{"/home/me"}}, nil)
	require.NoError(t, err)
	assert.Contains(t, strings.Join(f.invocation().Args, " "), "-o sftp.command=")
}

func TestCommandOptions(t *testing.T) {
	assert.Empty(t, CommandOptions("s3:host/bucket", map[string]string{"s3.storage-class": "GLACIER", "sftp.connections": "5"}, map[string]string{"AWS_ACCESS_KEY_ID": "a"}))
	assert.Equal(t, []string{"RCLONE.PROGRAM", "sftp.args"}, CommandOptions("sftp:h:/r", map[string]string{"sftp.args": "-o ProxyCommand=x", "RCLONE.PROGRAM": "/bin/sh"}, nil))
	assert.Len(t, CommandOptions("rclone::sftp,host=h,ssh='sh -c x':/r", nil, nil), 1)
	assert.Equal(t, []string{"RCLONE_CONFIG"}, CommandOptions("rclone:nas:backups", nil, map[string]string{"RCLONE_CONFIG": "/srv/data/restored.conf"}))
	assert.Empty(t, CommandOptions("rclone:nas:backups", nil, nil), "a remote of the default rclone.conf")
}
