package restic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepairIndex(t *testing.T) {
	f := newFake(t, scenario{Stdout: []string{"loading indexes...", "rebuilding index", "done"}})
	require.NoError(t, f.runner.RepairIndex(context.Background(), f.repo))
	assert.Equal(t, []string{"repair", "index"}, f.invocation().Args)

	locked := newFake(t, scenario{Exit: 11, Stderr: []string{"repository is already locked"}})
	assert.ErrorIs(t, locked.runner.RepairIndex(context.Background(), locked.repo), ErrLocked)
}

func TestRepairSnapshots(t *testing.T) {
	t.Run("some damaged", func(t *testing.T) {
		f := newFake(t, scenario{Stdout: []string{
			"", "snapshot 0c0a4b14 of [/home/me] at 2026-10-19 08:46:02 +0000 UTC by me@laptop",
			`  dir "/": not readable`, "removed empty snapshot 0c0a4b14", "", "modified 3 snapshots",
		}})
		n, err := f.runner.RepairSnapshots(context.Background(), f.repo)
		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, []string{"repair", "snapshots", "--forget"}, f.invocation().Args)
	})
	t.Run("none damaged", func(t *testing.T) {
		f := newFake(t, scenario{Stdout: []string{"snapshot 6001a440 of [/home/me]", "no snapshots were modified"}})
		n, err := f.runner.RepairSnapshots(context.Background(), f.repo)
		require.NoError(t, err)
		assert.Zero(t, n)
	})
	t.Run("failure", func(t *testing.T) {
		f := newFake(t, scenario{Exit: 1, Stderr: []string{"Fatal: unable to load index"}})
		_, err := f.runner.RepairSnapshots(context.Background(), f.repo)
		assert.ErrorContains(t, err, "unable to load index")
	})
}

func TestMigrations(t *testing.T) {
	f := newFake(t, scenario{Stdout: []string{"available migrations:", "  upgrade_repo_v2\tupgrade a repository to version 2"}})
	names, err := f.runner.Migrations(context.Background(), f.repo)
	require.NoError(t, err)
	assert.Equal(t, []string{"upgrade_repo_v2"}, names)
	assert.Equal(t, []string{"migrate"}, f.invocation().Args)

	none := newFake(t, scenario{Stdout: []string{"available migrations:", "no migrations found"}})
	names, err = none.runner.Migrations(context.Background(), none.repo)
	require.NoError(t, err)
	assert.Empty(t, names)

	wrong := newFake(t, scenario{Exit: 12})
	_, err = wrong.runner.Migrations(context.Background(), wrong.repo)
	assert.ErrorIs(t, err, ErrWrongPassword)
}

func TestMigrate(t *testing.T) {
	f := newFake(t, scenario{Stdout: []string{"checking repository integrity...", "no errors were found", "applying migration upgrade_repo_v2...", "migration upgrade_repo_v2: success"}})
	require.NoError(t, f.runner.Migrate(context.Background(), f.repo, "upgrade_repo_v2"))
	assert.Equal(t, []string{"migrate", "--", "upgrade_repo_v2"}, f.invocation().Args)

	// restic exits 0 for a migration it does not know
	unknown := newFake(t, scenario{Stdout: []string{"unknown migration bogus"}})
	assert.ErrorContains(t, unknown.runner.Migrate(context.Background(), unknown.repo, "bogus"), "bogus was not applied")

	failed := newFake(t, scenario{Exit: 1, Stderr: []string{"Fatal: repository contains errors"}})
	assert.ErrorContains(t, failed.runner.Migrate(context.Background(), failed.repo, "upgrade_repo_v2"), "repository contains errors")
}

func TestKeys(t *testing.T) {
	f := newFake(t, scenario{Stdout: []string{`[{"current":false,"id":"aaa","userName":"me","hostName":"laptop","created":"2026-10-01 10:00:00"},{"current":true,"id":"bbb","userName":"root","hostName":"nas","created":"2026-10-19 08:42:46"}]`}})
	keys, err := f.runner.Keys(context.Background(), f.repo)
	require.NoError(t, err)
	assert.Equal(t, []Key{
		{ID: "aaa", UserName: "me", HostName: "laptop", Created: "2026-10-01 10:00:00"},
		{ID: "bbb", Current: true, UserName: "root", HostName: "nas", Created: "2026-10-19 08:42:46"},
	}, keys)
	assert.Equal(t, []string{"key", "list", "--json"}, f.invocation().Args)
}

func TestAddKey(t *testing.T) {
	f := newFake(t, scenario{Stdout: []string{"saved new key as <Key of root@nas>"}})
	require.NoError(t, f.runner.AddKey(context.Background(), f.repo, "the new password"))
	inv := f.invocation()
	assert.Equal(t, []string{"key", "add", "--new-password-file"}, inv.Args[:3])
	// the new password goes through a file, never the command line or the environment
	assert.Equal(t, "the new password", inv.Files["--new-password-file"])
	for _, a := range inv.Args {
		assert.NotContains(t, a, "the new password")
	}
	for _, v := range inv.Env {
		assert.NotContains(t, v, "the new password")
	}
	assert.Equal(t, "repository-password", inv.Env["RESTIC_PASSWORD"])

	assert.ErrorContains(t, f.runner.AddKey(context.Background(), f.repo, ""), "empty")
}

func TestRemoveKey(t *testing.T) {
	f := newFake(t, scenario{})
	require.NoError(t, f.runner.RemoveKey(context.Background(), f.repo, "aaa"))
	assert.Equal(t, []string{"key", "remove", "--", "aaa"}, f.invocation().Args)

	current := newFake(t, scenario{Exit: 1, Stderr: []string{"Fatal: refusing to remove key currently used to access repository"}})
	assert.ErrorContains(t, current.runner.RemoveKey(context.Background(), current.repo, "bbb"), "refusing")
}
