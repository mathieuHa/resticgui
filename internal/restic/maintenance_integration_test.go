//go:build integration

package restic

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationRepair damages a repository the way a lost file on the
// storage would, then repairs it as the troubleshooting guide says.
func TestIntegrationRepair(t *testing.T) {
	ctx := context.Background()
	r := realRunner(t)
	repo := &Repo{Location: filepath.Join(t.TempDir(), "repo"), Password: "integration-password"}
	require.NoError(t, r.Init(ctx, repo))
	_, err := r.Backup(ctx, repo, BackupOptions{Paths: []string{writeTree(t)}}, nil)
	require.NoError(t, err)

	// nothing to repair yet
	require.NoError(t, r.RepairIndex(ctx, repo))
	n, err := r.RepairSnapshots(ctx, repo)
	require.NoError(t, err)
	assert.Zero(t, n)

	// lose every pack file
	packs, err := filepath.Glob(filepath.Join(repo.Location, "data", "*", "*"))
	require.NoError(t, err)
	require.NotEmpty(t, packs)
	for _, p := range packs {
		require.NoError(t, os.Remove(p))
	}
	_, err = r.Check(ctx, repo, CheckOptions{})
	require.Error(t, err, "the check finds the damage")

	require.NoError(t, r.RepairIndex(ctx, repo))
	n, err = r.RepairSnapshots(ctx, repo)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	summary, err := r.Check(ctx, repo, CheckOptions{ReadDataSubset: "100%"})
	require.NoError(t, err, "repaired")
	require.NotNil(t, summary)
	assert.Zero(t, summary.NumErrors)
}

func TestIntegrationMigrate(t *testing.T) {
	ctx := context.Background()
	r := realRunner(t)
	repo := &Repo{Location: filepath.Join(t.TempDir(), "repo"), Password: "integration-password"}
	require.NoError(t, r.run(ctx, repo, call{command: "init", args: []string{"--repository-version", "1"}}))

	names, err := r.Migrations(ctx, repo)
	require.NoError(t, err)
	assert.Contains(t, names, "upgrade_repo_v2")
	require.NoError(t, r.Migrate(ctx, repo, "upgrade_repo_v2"))

	names, err = r.Migrations(ctx, repo)
	require.NoError(t, err)
	assert.Empty(t, names, "version 2 is the latest")
	assert.ErrorContains(t, r.Migrate(ctx, repo, "no_such_migration"), "was not applied")
}

func TestIntegrationKeys(t *testing.T) {
	ctx := context.Background()
	r := realRunner(t)
	repo := &Repo{Location: filepath.Join(t.TempDir(), "repo"), Password: "old password"}
	require.NoError(t, r.Init(ctx, repo))

	keys, err := r.Keys(ctx, repo)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	old := keys[0]
	assert.True(t, old.Current)

	require.NoError(t, r.AddKey(ctx, repo, "new password"))
	next := *repo
	next.Password = "new password"
	keys, err = r.Keys(ctx, &next)
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.Error(t, r.RemoveKey(ctx, repo, old.ID), "restic keeps the key in use")
	require.NoError(t, r.RemoveKey(ctx, &next, old.ID))

	_, err = r.Keys(ctx, repo)
	assert.ErrorIs(t, err, ErrWrongPassword, "the old password no longer opens it")
	keys, err = r.Keys(ctx, &next)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}
