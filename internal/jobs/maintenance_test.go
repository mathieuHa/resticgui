package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

func TestCheckReadsADataSubset(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	e.restic.check = &restic.CheckSummary{}

	job, err := e.manager.EnqueueCheck(ctx, repo.ID, CheckParams{ReadDataSubset: "10%"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"read_data_subset":"10%"}`, job.Params)
	e.waitStatus(t, job.ID, StatusSucceeded)

	// a scheduled check reads what its schedule says
	_, err = e.store.UpsertSchedule(ctx, db.UpsertScheduleParams{RepositoryID: repo.ID, Job: "check", Cron: "@daily", Enabled: true, ReadDataSubset: "1/5"})
	require.NoError(t, err)
	scheduled, err := e.manager.Enqueue(ctx, repo.ID, 0, KindCheck, TriggerSchedule)
	require.NoError(t, err)
	e.waitStatus(t, scheduled.ID, StatusSucceeded)

	// a manual check from Enqueue reads no data
	manual, err := e.manager.Enqueue(ctx, repo.ID, 0, KindCheck, TriggerManual)
	require.NoError(t, err)
	e.waitStatus(t, manual.ID, StatusSucceeded)

	var subsets []string
	for _, c := range e.restic.Calls() {
		subsets = append(subsets, c.Opts.(restic.CheckOptions).ReadDataSubset)
	}
	assert.Equal(t, []string{"10%", "1/5", ""}, subsets)
}

func TestValidSubset(t *testing.T) {
	for _, ok := range []string{"", "10%", "2.5%", "1/5", "500M", "2G", "100"} {
		assert.NoError(t, ValidSubset(ok), ok)
	}
	for _, bad := range []string{"0/5", "-1%", "10 %", "--read-data", "1/5; rm", "five"} {
		assert.Error(t, ValidSubset(bad), bad)
	}
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	_, err := e.manager.EnqueueCheck(context.Background(), repo.ID, CheckParams{ReadDataSubset: "--help"})
	assert.ErrorContains(t, err, "read_data_subset")
	assert.Empty(t, e.restic.Calls())
}

func TestRepairJobs(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	e.restic.modified = 2

	index, err := e.manager.Enqueue(ctx, repo.ID, 0, KindRepairIndex, TriggerManual)
	require.NoError(t, err)
	e.waitStatus(t, index.ID, StatusSucceeded)
	snaps, err := e.manager.Enqueue(ctx, repo.ID, 0, KindRepairSnapshots, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, snaps.ID, StatusSucceeded)
	assert.JSONEq(t, `{"modified":2}`, done.Summary)

	e.restic.run = func(_ context.Context, c fakeCall) error {
		return &restic.Error{Command: "repair", ExitCode: 11, Stderr: []string{"repository is already locked"}}
	}
	locked, err := e.manager.Enqueue(ctx, repo.ID, 0, KindRepairIndex, TriggerManual)
	require.NoError(t, err)
	failed := e.waitStatus(t, locked.ID, StatusFailed)
	assert.Equal(t, int64(11), *failed.ExitCode)
}

func TestMigrateJob(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")

	e.restic.migrations = []string{"upgrade_repo_v2"}
	job, err := e.manager.Enqueue(ctx, repo.ID, 0, KindMigrate, TriggerManual)
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusSucceeded)
	assert.JSONEq(t, `{"applied":["upgrade_repo_v2"]}`, done.Summary)

	// a repository already at version 2 has nothing to apply
	e.restic.migrations = nil
	job, err = e.manager.Enqueue(ctx, repo.ID, 0, KindMigrate, TriggerManual)
	require.NoError(t, err)
	done = e.waitStatus(t, job.ID, StatusSucceeded)
	assert.JSONEq(t, `{"applied":[]}`, done.Summary)

	var calls []string
	for _, c := range e.restic.Calls() {
		calls = append(calls, c.Kind)
	}
	assert.Equal(t, []string{"migrations", "migrate", "migrations"}, calls)
	assert.Equal(t, "upgrade_repo_v2", e.restic.Calls()[1].Opts)
}

// password reads the repository's saved password.
func (e *env) password(t *testing.T, id int64) string {
	t.Helper()
	r, err := e.store.GetRepository(context.Background(), id)
	require.NoError(t, err)
	sec, err := store.OpenSecrets(e.box, r.Secrets)
	require.NoError(t, err)
	return sec.Password
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	e.restic.keys = map[string]string{"pw-r": "old", "someone else's": "other"}

	job, err := e.manager.EnqueuePasswordChange(ctx, repo.ID, "a new password")
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusSucceeded)
	assert.Equal(t, "a new password", e.password(t, repo.ID))
	assert.Equal(t, map[string]string{"a new password": "key-a new password", "someone else's": "other"}, e.restic.keys,
		"the old key is removed, other keys stay")
	// the old key is removed with the new password
	calls := e.restic.Calls()
	last := calls[len(calls)-1]
	assert.Equal(t, "key remove", last.Kind)
	assert.Equal(t, "old", last.Opts)
	assert.Equal(t, "a new password", last.Repo.Password)

	// the password is in no column of the job
	raw, err := json.Marshal(done)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "a new password")

	_, err = e.manager.EnqueuePasswordChange(ctx, repo.ID, "")
	assert.Error(t, err)
	_, err = e.manager.Enqueue(ctx, repo.ID, 0, KindChangePassword, TriggerManual)
	assert.ErrorIs(t, err, ErrUnknownKind)
}

func TestChangePasswordKeepsTheOldPasswordWhenTheNewKeyFails(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	e.restic.keys = map[string]string{"pw-r": "old"}
	e.restic.run = func(_ context.Context, c fakeCall) error {
		if c.Kind == "key add" {
			return errors.New("restic key: exit code 1: Fatal: unable to save key")
		}
		return nil
	}
	job, err := e.manager.EnqueuePasswordChange(ctx, repo.ID, "a new password")
	require.NoError(t, err)
	done := e.waitStatus(t, job.ID, StatusFailed)
	assert.Contains(t, done.Error, "unable to save key")
	assert.Equal(t, "pw-r", e.password(t, repo.ID))
	assert.Equal(t, map[string]string{"pw-r": "old"}, e.restic.keys)
}

func TestChangePasswordWhenTheOldKeyStays(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	repo := e.repository(t, "r")
	e.restic.keys = map[string]string{"pw-r": "old"}
	e.restic.run = func(_ context.Context, c fakeCall) error {
		if c.Kind == "key remove" {
			return errors.New("restic key: exit code 1: Fatal: unable to remove key")
		}
		return nil
	}
	job, err := e.manager.EnqueuePasswordChange(ctx, repo.ID, "a new password")
	require.NoError(t, err)
	// the new password is saved and works: the job succeeds and says the old one still works
	done := e.waitStatus(t, job.ID, StatusSucceeded)
	assert.Contains(t, done.Error, "the old one still opens the repository")
	assert.Equal(t, "a new password", e.password(t, repo.ID))
}

func TestChangePasswordIsNotCancelled(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t, Options{})
	g := newGate()
	repo := e.repository(t, "r")
	e.restic.keys = map[string]string{"pw-r": "old"}
	e.restic.run = func(ctx context.Context, c fakeCall) error {
		if c.Kind == "key add" {
			return g.run(ctx, c)
		}
		return nil
	}
	job, err := e.manager.EnqueuePasswordChange(ctx, repo.ID, "a new password")
	require.NoError(t, err)
	g.waitStarted(t)
	require.NoError(t, e.manager.Cancel(ctx, job.ID))
	g.release <- struct{}{}
	e.waitStatus(t, job.ID, StatusSucceeded)
	assert.Equal(t, "a new password", e.password(t, repo.ID))
}
