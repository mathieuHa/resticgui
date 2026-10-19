//go:build integration

package jobs

// The job manager driving a real restic binary against a local repository:
//
//	RESTIC_BIN=/path/to/restic go test -tags integration ./internal/jobs/

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

func TestIntegrationJobs(t *testing.T) {
	bin := os.Getenv("RESTIC_BIN")
	if bin == "" {
		bin = "restic"
	}
	ctx := context.Background()
	e := newEnv(t, Options{})
	e.manager = New(e.store, e.box, restic.NewRunner(bin), e.events, Options{})
	t.Cleanup(e.manager.Close)

	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0o644))
	sealed, err := store.SealSecrets(e.box, store.RepositorySecrets{Password: "integration"})
	require.NoError(t, err)
	repo, err := e.store.CreateRepository(ctx, db.CreateRepositoryParams{
		Name: "local", Backend: "local", Location: filepath.Join(t.TempDir(), "repo"), Options: "{}",
		Secrets: sealed, Host: "integration",
	})
	require.NoError(t, err)
	e.addPlan(t, repo.ID, "Default", store.PlanOptions{})
	e.addPath(t, repo.ID, src, "include")
	require.NoError(t, e.store.UpsertPlanRetention(ctx, db.UpsertPlanRetentionParams{PlanID: e.plan(t, repo.ID).ID, KeepLast: 1, Prune: true}))

	run := func(kind Kind) db.Job {
		t.Helper()
		job, err := e.manager.Enqueue(ctx, repo.ID, 0, kind, TriggerManual)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			job, err = e.store.GetJob(ctx, job.ID)
			require.NoError(t, err)
			return job.FinishedAt != nil
		}, 2*time.Minute, 50*time.Millisecond)
		require.Equal(t, string(StatusSucceeded), job.Status, "%s: %s\n%s", kind, job.Error, job.LogExcerpt)
		return job
	}

	run(KindInit)
	first := run(KindBackup)
	assert.NotEmpty(t, first.SnapshotID)
	require.NoError(t, os.WriteFile(filepath.Join(src, "b.txt"), []byte("world"), 0o644))
	second := run(KindBackup)
	assert.NotEqual(t, first.SnapshotID, second.SnapshotID)

	forget := run(KindForget)
	var summary ForgetSummary
	require.NoError(t, json.Unmarshal([]byte(forget.Summary), &summary))
	assert.Equal(t, ForgetSummary{Kept: 1, Removed: 1, Pruned: true}, summary)
	run(KindCheck)

	// the maintenance tools
	check, err := e.manager.EnqueueCheck(ctx, repo.ID, CheckParams{ReadDataSubset: "50%"})
	require.NoError(t, err)
	require.Equal(t, string(StatusSucceeded), e.waitDone(t, check.ID).Status)
	run(KindUnlock)
	run(KindRepairIndex)
	assert.JSONEq(t, `{"modified":0}`, run(KindRepairSnapshots).Summary)
	assert.JSONEq(t, `{"applied":[]}`, run(KindMigrate).Summary, "restic makes version 2 repositories")

	change, err := e.manager.EnqueuePasswordChange(ctx, repo.ID, "a new password")
	require.NoError(t, err)
	done := e.waitDone(t, change.ID)
	require.Equal(t, string(StatusSucceeded), done.Status, "%s\n%s", done.Error, done.LogExcerpt)
	assert.Empty(t, done.Error)
	assert.Equal(t, "a new password", e.password(t, repo.ID))
	// the jobs go on with the new password, and the old one opens nothing
	run(KindBackup)
	stored, err := e.store.GetRepository(ctx, repo.ID)
	require.NoError(t, err)
	target, err := ResticRepo(e.box, stored)
	require.NoError(t, err)
	target.Password = "integration"
	_, err = restic.NewRunner(bin).Keys(ctx, target)
	assert.ErrorIs(t, err, restic.ErrWrongPassword)
}

// waitDone waits for a job to end, for the slow jobs of real restic.
func (e *env) waitDone(t *testing.T, id int64) db.Job {
	t.Helper()
	var job db.Job
	require.Eventually(t, func() bool {
		var err error
		job, err = e.store.GetJob(context.Background(), id)
		require.NoError(t, err)
		return job.FinishedAt != nil
	}, 2*time.Minute, 50*time.Millisecond)
	return job
}

// Two plans back up into one repository with real restic, and the
// retention of one keeps the other's snapshots.
func TestIntegrationTwoPlans(t *testing.T) {
	bin := os.Getenv("RESTIC_BIN")
	if bin == "" {
		bin = "restic"
	}
	ctx := context.Background()
	e := newEnv(t, Options{})
	runner := restic.NewRunner(bin)
	e.manager = New(e.store, e.box, runner, e.events, Options{})
	t.Cleanup(e.manager.Close)

	photos, docs := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(photos, "a.jpg"), []byte("photo"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(docs, "a.txt"), []byte("doc"), 0o644))
	sealed, err := store.SealSecrets(e.box, store.RepositorySecrets{Password: "integration"})
	require.NoError(t, err)
	repo, err := e.store.CreateRepository(ctx, db.CreateRepositoryParams{
		Name: "local", Backend: "local", Location: filepath.Join(t.TempDir(), "repo"), Options: "{}",
		Secrets: sealed, Host: "integration",
	})
	require.NoError(t, err)
	var plans []db.Plan
	for _, p := range []store.Plan{
		{Name: "Photos", Include: []string{photos}, Tags: []string{"camera"}},
		{Name: "Documents", Include: []string{docs}, Options: store.PlanOptions{Compression: "max", ExcludeCaches: true}, Retention: &db.PlanRetention{KeepLast: 1}},
	} {
		require.NoError(t, e.store.InTx(ctx, func(q *db.Queries) error {
			plan, err := store.SavePlan(ctx, q, repo.ID, 0, p)
			plans = append(plans, plan)
			return err
		}))
	}

	run := func(planID int64, kind Kind) db.Job {
		t.Helper()
		job, err := e.manager.Enqueue(ctx, repo.ID, planID, kind, TriggerManual)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			job, err = e.store.GetJob(ctx, job.ID)
			require.NoError(t, err)
			return job.FinishedAt != nil
		}, 2*time.Minute, 50*time.Millisecond)
		require.Equal(t, string(StatusSucceeded), job.Status, "%s: %s\n%s", kind, job.Error, job.LogExcerpt)
		return job
	}
	run(0, KindInit)
	for range 2 {
		run(plans[0].ID, KindBackup)
		run(plans[1].ID, KindBackup)
	}
	forget := run(plans[1].ID, KindForget)
	var summary ForgetSummary
	require.NoError(t, json.Unmarshal([]byte(forget.Summary), &summary))
	assert.Equal(t, ForgetSummary{Kept: 1, Removed: 1}, summary)

	target, err := ResticRepo(e.box, repo)
	require.NoError(t, err)
	snaps, err := runner.Snapshots(ctx, target, restic.SnapshotFilter{})
	require.NoError(t, err)
	byPlan := map[string]int{}
	for _, s := range snaps {
		require.True(t, OfRepository(repo.ID, s.Tags), "%v", s.Tags)
		byPlan[s.Tags[0]]++
		if s.Tags[0] == plans[0].Tag {
			assert.Equal(t, []string{plans[0].Tag, "camera"}, s.Tags)
			assert.Equal(t, []string{photos}, s.Paths)
		}
	}
	assert.Equal(t, map[string]int{plans[0].Tag: 2, plans[1].Tag: 1}, byPlan, "the photos are all kept")
}
