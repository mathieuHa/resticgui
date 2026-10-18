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
		Secrets: sealed, Host: "integration", Compression: "auto",
	})
	require.NoError(t, err)
	_, err = e.store.AddPath(ctx, db.AddPathParams{RepositoryID: repo.ID, Path: src, Kind: "include"})
	require.NoError(t, err)
	_, err = e.store.UpsertRetention(ctx, db.UpsertRetentionParams{RepositoryID: repo.ID, KeepLast: 1, Prune: true})
	require.NoError(t, err)

	run := func(kind Kind) db.Job {
		t.Helper()
		job, err := e.manager.Enqueue(ctx, repo.ID, kind, TriggerManual)
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
}
