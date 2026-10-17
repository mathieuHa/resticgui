package restic

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		f := newFake(t, scenario{Stdout: []string{`{"message_type":"version","version":"0.18.1","go_version":"go1.25.1","go_os":"linux","go_arch":"amd64"}`}})
		f.runner.Env = append(f.runner.Env, "FAKE_RESTIC_SCENARIO="+f.repo.Env["FAKE_RESTIC_SCENARIO"])
		v, err := f.runner.CheckVersion(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "0.18.1", v)
		assert.Equal(t, []string{"version", "--json"}, f.invocation().Args)
	})
	t.Run("text from an old restic", func(t *testing.T) {
		f := newFake(t, scenario{Stdout: []string{"restic 0.16.4 compiled with go1.21.6 on linux/amd64"}})
		f.runner.Env = append(f.runner.Env, "FAKE_RESTIC_SCENARIO="+f.repo.Env["FAKE_RESTIC_SCENARIO"])
		v, err := f.runner.CheckVersion(context.Background())
		assert.Equal(t, "0.16.4", v)
		assert.ErrorContains(t, err, "too old")
	})
}

func TestCompareVersions(t *testing.T) {
	assert.Equal(t, 0, compareVersions("0.17.0", "0.17.0"))
	assert.Equal(t, -1, compareVersions("0.16.10", "0.17.0"))
	assert.Equal(t, 1, compareVersions("0.17.10", "0.17.9"))
	assert.Equal(t, 1, compareVersions("1.0.0", "0.17.0"))
}

func TestExists(t *testing.T) {
	for _, tt := range []struct {
		exit    int
		want    bool
		wantErr error
	}{{0, true, nil}, {10, false, nil}, {12, false, ErrWrongPassword}} {
		f := newFake(t, scenario{Exit: tt.exit})
		got, err := f.runner.Exists(context.Background(), f.repo)
		assert.Equal(t, tt.want, got)
		if tt.wantErr != nil {
			assert.ErrorIs(t, err, tt.wantErr)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, []string{"cat", "config", "--json"}, f.invocation().Args[:3])
	}
}

func TestBackup(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/backup.jsonl"})
	var progress []BackupStatus
	res, err := f.runner.Backup(context.Background(), f.repo, BackupOptions{
		Paths:             []string{"/home/me/docs", "/home/me/My Pictures"},
		Excludes:          []string{"/home/me/docs/tmp", "*.iso"},
		Tags:              []string{"resticgui:1"},
		Host:              "laptop",
		ExcludeLargerThan: "2G",
		Compression:       "max",
	}, func(s BackupStatus) { progress = append(progress, s) })
	require.NoError(t, err)

	require.Len(t, progress, 2)
	assert.Equal(t, uint64(51), progress[1].TotalFiles)
	assert.InDelta(t, 1.0, progress[1].PercentDone, 0.0001)
	assert.Equal(t, []string{"/src/a/f18"}, progress[0].CurrentFiles)

	require.NotNil(t, res.Summary)
	assert.Equal(t, "f0feba17d45589dc1082af60dd77a6cf55a7e4fdbcc4a0fcfcd0c06080b50f1e", res.Summary.SnapshotID)
	assert.Equal(t, uint64(51), res.Summary.FilesNew)
	assert.Equal(t, uint64(1018580), res.Summary.DataAdded)
	assert.Empty(t, res.Errors)

	inv := f.invocation()
	assert.Equal(t, "/home/me/docs\n/home/me/My Pictures\n", inv.Files["--files-from-verbatim"])
	assert.Equal(t, "/home/me/docs/tmp\n*.iso\n", inv.Files["--exclude-file"])
	joined := strings.Join(inv.Args, " ")
	for _, want := range []string{"backup --json", "--tag resticgui:1", "--host laptop", "--exclude-larger-than 2G", "--compression max"} {
		assert.Contains(t, joined, want)
	}
	assert.NotContains(t, joined, "--dry-run")
	assert.NotContains(t, inv.Args, "")
}

func TestBackupDryRunAndEmptyPaths(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/backup.jsonl"})
	_, err := f.runner.Backup(context.Background(), f.repo, BackupOptions{Paths: []string{"/x"}, DryRun: true}, nil)
	require.NoError(t, err)
	assert.Contains(t, f.invocation().Args, "--dry-run")
	assert.NotContains(t, f.invocation().Args, "--exclude-file")

	_, err = f.runner.Backup(context.Background(), f.repo, BackupOptions{}, nil)
	assert.ErrorContains(t, err, "no paths")
}

func TestBackupIncomplete(t *testing.T) {
	f := newFake(t, scenario{
		StdoutFile: "testdata/backup-partial.jsonl", Exit: 3,
		Stderr: []string{`{"message_type":"exit_error","code":3,"message":"Warning: at least one source file could not be read"}`},
	})
	res, err := f.runner.Backup(context.Background(), f.repo, BackupOptions{Paths: []string{"/home/me/src"}}, nil)
	assert.ErrorIs(t, err, ErrIncomplete)
	require.NotNil(t, res, "the partial snapshot is still reported")
	assert.Equal(t, "9a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f9", res.Summary.SnapshotID)
	assert.Equal(t, []BackupError{{Message: "open /home/me/src/secret: permission denied", During: "archival", Item: "/home/me/src/secret"}}, res.Errors)
}

func TestBackupFailureWithoutSummary(t *testing.T) {
	f := newFake(t, scenario{Exit: 1, Stderr: []string{`{"message_type":"exit_error","code":1,"message":"Fatal: all source directories/files do not exist"}`}})
	res, err := f.runner.Backup(context.Background(), f.repo, BackupOptions{Paths: []string{"/nonexistent"}}, nil)
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "all source directories/files do not exist")

	f = newFake(t, scenario{Stdout: []string{`{"message_type":"status","percent_done":1}`}})
	_, err = f.runner.Backup(context.Background(), f.repo, BackupOptions{Paths: []string{"/x"}}, nil)
	assert.ErrorContains(t, err, "no summary")
}

func TestSnapshots(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/snapshots.json"})
	snaps, err := f.runner.Snapshots(context.Background(), f.repo, SnapshotFilter{Host: "h1", Tags: []string{"resticgui:1", "daily"}})
	require.NoError(t, err)
	require.Len(t, snaps, 2)
	assert.Equal(t, "f0feba17", snaps[0].ShortID)
	assert.Equal(t, []string{"/home/me/src"}, snaps[0].Paths)
	assert.Equal(t, []string{"resticgui:1"}, snaps[0].Tags)
	require.NotNil(t, snaps[0].Summary)
	assert.Equal(t, uint64(51), snaps[0].Summary.FilesNew)
	assert.Equal(t, "f0feba17d45589dc1082af60dd77a6cf55a7e4fdbcc4a0fcfcd0c06080b50f1e", snaps[1].Parent)
	assert.Equal(t, []string{"snapshots", "--json", "--host", "h1", "--tag", "resticgui:1,daily"}, f.invocation().Args[:6])

	empty := newFake(t, scenario{Stdout: []string{"[]"}})
	snaps, err = empty.runner.Snapshots(context.Background(), empty.repo, SnapshotFilter{})
	require.NoError(t, err)
	assert.NotNil(t, snaps)
	assert.Empty(t, snaps)
}

func TestForget(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/forget.json", Stderr: []string{"loading indexes...", "to repack: 0 blobs / 0 B"}})
	groups, err := f.runner.Forget(context.Background(), f.repo, ForgetOptions{
		Policy:  Policy{KeepLast: 3, KeepHourly: 24, KeepDaily: 7, KeepWeekly: 4, KeepMonthly: 12, KeepYearly: 2, KeepWithin: "30d"},
		Filter:  SnapshotFilter{Tags: []string{"resticgui:1"}},
		GroupBy: "host,tags",
		Prune:   true,
		DryRun:  true,
	})
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "h1", groups[0].Host)
	assert.Len(t, groups[0].Keep, 1)
	assert.Len(t, groups[0].Remove, 1)
	assert.Equal(t, []string{"last snapshot"}, groups[0].Reasons[0].Matches)

	assert.Equal(t, []string{
		"forget", "--json",
		"--keep-last", "3", "--keep-hourly", "24", "--keep-daily", "7", "--keep-weekly", "4",
		"--keep-monthly", "12", "--keep-yearly", "2", "--keep-within", "30d",
		"--tag", "resticgui:1", "--group-by", "host,tags", "--prune", "--dry-run",
	}, f.invocation().Args)
}

func TestForgetRefusesEmptyPolicy(t *testing.T) {
	f := newFake(t, scenario{})
	_, err := f.runner.Forget(context.Background(), f.repo, ForgetOptions{Prune: true})
	assert.ErrorIs(t, err, ErrEmptyPolicy)
}

func TestForgetSnapshots(t *testing.T) {
	f := newFake(t, scenario{})
	require.NoError(t, f.runner.ForgetSnapshots(context.Background(), f.repo, []string{"f0feba17", "-not-a-flag"}, true))
	assert.Equal(t, []string{"forget", "--json", "--prune", "--", "f0feba17", "-not-a-flag"}, f.invocation().Args)
	assert.Error(t, f.runner.ForgetSnapshots(context.Background(), f.repo, nil, false))
}

func TestCheck(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/check.json"})
	s, err := f.runner.Check(context.Background(), f.repo, CheckOptions{ReadDataSubset: "10%"})
	require.NoError(t, err)
	assert.Equal(t, &CheckSummary{}, s)
	assert.Equal(t, []string{"check", "--json", "--read-data-subset", "10%"}, f.invocation().Args[:4])

	broken := newFake(t, scenario{Exit: 1, Stdout: []string{`{"message_type":"summary","num_errors":2,"broken_packs":["abc"],"suggest_repair_index":true,"suggest_prune":false}`}})
	s, err = broken.runner.Check(context.Background(), broken.repo, CheckOptions{})
	assert.Error(t, err)
	require.NotNil(t, s)
	assert.Equal(t, 2, s.NumErrors)
	assert.Equal(t, []string{"abc"}, s.BrokenPacks)
}

func TestRestore(t *testing.T) {
	f := newFake(t, scenario{Stdout: []string{
		`{"message_type":"status","seconds_elapsed":1,"percent_done":0.5,"total_files":55,"files_restored":20,"total_bytes":1000011,"bytes_restored":500000}`,
	}, StdoutFile: "testdata/restore.jsonl"})
	var progress []RestoreStatus
	s, err := f.runner.Restore(context.Background(), f.repo, RestoreOptions{
		SnapshotID: "latest", Target: "/tmp/restore", Include: []string{"/home/me/docs"},
	}, func(p RestoreStatus) { progress = append(progress, p) })
	require.NoError(t, err)
	require.Len(t, progress, 1)
	assert.Equal(t, uint64(20), progress[0].FilesRestored)
	assert.Equal(t, uint64(55), s.FilesRestored)
	assert.Equal(t, []string{"restore", "--json", "--target", "/tmp/restore", "--include", "/home/me/docs", "--", "latest"}, f.invocation().Args[:8])

	_, err = f.runner.Restore(context.Background(), f.repo, RestoreOptions{SnapshotID: "latest"}, nil)
	assert.ErrorContains(t, err, "required")
}

func TestLs(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/ls.jsonl"})
	nodes, err := f.runner.Ls(context.Background(), f.repo, "b805b5ae", "/src")
	require.NoError(t, err)
	require.Len(t, nodes, 5, "the snapshot line is skipped")
	assert.Equal(t, Node{
		Name: "src", Type: "dir", Path: "/src", Mode: 2147484141, Permissions: "drwxr-xr-x",
		MTime: nodes[0].MTime,
	}, nodes[0])
	assert.Equal(t, []string{"ls", "--json", "--", "b805b5ae", "/src"}, f.invocation().Args[:5])
}

func TestFind(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/find.json"})
	results, err := f.runner.Find(context.Background(), f.repo, "latest", "f1*")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 11, results[0].Hits)
	assert.Equal(t, "/src/a/f1", results[0].Matches[0].Path)
	assert.Equal(t, uint64(20000), results[0].Matches[0].Size)
	assert.Equal(t, []string{"find", "--json", "--snapshot", "latest", "--", "f1*"}, f.invocation().Args[:6])
}

func TestStats(t *testing.T) {
	f := newFake(t, scenario{StdoutFile: "testdata/stats-raw.json"})
	s, err := f.runner.Stats(context.Background(), f.repo, "raw-data", "latest")
	require.NoError(t, err)
	assert.Equal(t, uint64(1012642), s.TotalSize)
	assert.Equal(t, 3, s.SnapshotsCount)
	assert.InDelta(t, 1.0445, s.CompressionRatio, 0.001)
	assert.Equal(t, []string{"stats", "--json", "--mode", "raw-data", "--", "latest"}, f.invocation().Args[:6])
}

func TestDump(t *testing.T) {
	f := newFake(t, scenario{Stdout: []string{"file content"}})
	var out bytes.Buffer
	require.NoError(t, f.runner.Dump(context.Background(), f.repo, "latest", "/home/me/docs", "zip", &out))
	assert.Equal(t, "file content\n", out.String())
	assert.Equal(t, []string{"dump", "--archive", "zip", "--", "latest", "/home/me/docs"}, f.invocation().Args[:6])
}
