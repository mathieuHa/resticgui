package jobs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/hooks"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// addHook stores a command hook that appends its event to trace.
func (e *env) addHook(t *testing.T, repositoryID int64, event, command string, timeout int64) {
	t.Helper()
	if timeout == 0 {
		timeout = 30
	}
	_, err := e.store.CreateHook(context.Background(), db.CreateHookParams{
		RepositoryID: repositoryID, Event: event, Kind: hooks.Command, Command: command, TimeoutSeconds: timeout,
	})
	require.NoError(t, err)
}

func readTrace(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	return strings.Fields(string(raw))
}

func TestHooksRunAroundTheBackupInOrder(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{Hooks: hooks.Runner{Commands: true}})
	repo := e.repository(t, "photos")
	trace := filepath.Join(t.TempDir(), "trace")
	e.addHook(t, repo.ID, hooks.Before, "echo before-1 >> "+trace+"; echo dumping the database", 0)
	e.addHook(t, repo.ID, hooks.Before, "echo before-2 >> "+trace, 0)
	e.addHook(t, repo.ID, hooks.Success, `echo "success-$RESTICGUI_SNAPSHOT_ID" >> `+trace, 0)
	e.addHook(t, repo.ID, hooks.Failure, "echo failure >> "+trace, 0)
	e.restic.backup = &restic.BackupResult{Summary: &restic.BackupSummary{SnapshotID: "abc123"}}
	e.restic.run = func(ctx context.Context, c fakeCall) error {
		// the backup sees what the before hooks did
		assert.Equal(t, []string{"before-1", "before-2"}, readTrace(t, trace))
		return nil
	}

	job, err := e.manager.Enqueue(ctx, repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusSucceeded)
	assert.Equal(t, []string{"before-1", "before-2", "success-abc123"}, readTrace(t, trace))
	assert.Equal(t, "[before hook 1] dumping the database", done.LogExcerpt)
	assert.Empty(t, done.Error)
}

func TestAFailingBeforeHookStopsTheBackup(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{Hooks: hooks.Runner{Commands: true}})
	repo := e.repository(t, "photos")
	trace := filepath.Join(t.TempDir(), "trace")
	e.addHook(t, repo.ID, hooks.Before, "echo database is down; exit 2", 0)
	e.addHook(t, repo.ID, hooks.Before, "echo never >> "+trace, 0)
	e.addHook(t, repo.ID, hooks.Failure, `echo "failure:$RESTICGUI_ERROR" | tr ' ' _ >> `+trace, 0)

	job, err := e.manager.Enqueue(ctx, repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusFailed)
	assert.Empty(t, e.restic.Calls(), "no backup after a failed before hook")
	assert.Equal(t, "before hook 1 failed: exited with code 2", done.Error)
	assert.Equal(t, "[before hook 1] database is down\n[before hook 1] failed: exited with code 2", done.LogExcerpt)
	assert.Equal(t, []string{"failure:before_hook_1_failed:_exited_with_code_2"}, readTrace(t, trace))
}

func TestFailureHooksRunWhenTheBackupFails(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{Hooks: hooks.Runner{Commands: true}})
	repo := e.repository(t, "photos")
	trace := filepath.Join(t.TempDir(), "trace")
	e.addHook(t, repo.ID, hooks.Success, "echo success >> "+trace, 0)
	e.addHook(t, repo.ID, hooks.Failure, `echo "$RESTICGUI_STATUS" >> `+trace, 0)
	e.restic.run = func(ctx context.Context, c fakeCall) error { return errors.New("repository is locked") }

	job, err := e.manager.Enqueue(ctx, repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusFailed)
	assert.Equal(t, "repository is locked", done.Error)
	assert.Equal(t, []string{"failed"}, readTrace(t, trace))
}

func TestAFailingSuccessHookKeepsTheSnapshot(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{Hooks: hooks.Runner{Commands: true}})
	repo := e.repository(t, "photos")
	e.addHook(t, repo.ID, hooks.Success, "sleep 10", 1)
	e.restic.backup = &restic.BackupResult{Summary: &restic.BackupSummary{SnapshotID: "abc123"}}

	start := time.Now()
	job, err := e.manager.Enqueue(ctx, repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusSucceeded)
	assert.Less(t, time.Since(start), 8*time.Second, "the hook's timeout stopped it")
	assert.Equal(t, "abc123", done.SnapshotID)
	assert.Equal(t, "success hook 1 failed: timed out after 1s", done.Error)
}

func TestCommandHooksNeedTheConfiguration(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "photos")
	e.addHook(t, repo.ID, hooks.Before, "true", 0)

	job, err := e.manager.Enqueue(ctx, repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusFailed)
	assert.Contains(t, done.Error, "commands are turned off")
	assert.Empty(t, e.restic.Calls())
}

func TestStatsAreRefreshedAfterABackup(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "photos")
	e.restic.stats = restic.Stats{TotalSize: 4096, SnapshotsCount: 3}

	job, err := e.manager.Enqueue(ctx, repo.ID, KindBackup, TriggerManual)
	require.NoError(t, err)
	e.waitStatus(t, job.ID, StatusSucceeded)
	require.Eventually(t, func() bool {
		stats, err := e.store.ListRepositoryStats(ctx)
		require.NoError(t, err)
		return len(stats) == 1 && stats[0].Snapshots == 3 && stats[0].TotalSize == 4096
	}, 5*time.Second, 10*time.Millisecond)
}
