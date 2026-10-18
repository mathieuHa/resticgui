//go:build integration

package restic

// Integration tests against a real restic binary:
//
//	RESTIC_BIN=/path/to/restic go test -tags integration ./internal/restic/
//
// The S3 variant runs when RESTIC_TEST_S3_LOCATION is set (for example
// s3:http://127.0.0.1:9000/resticgui-test) with AWS_ACCESS_KEY_ID and
// AWS_SECRET_ACCESS_KEY for that server; CI points it at a Garage container
// configured by testdata/garage.toml.

import (
	"bytes"
	"context"
	"errors"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func realRunner(t *testing.T) *Runner {
	t.Helper()
	bin := os.Getenv("RESTIC_BIN")
	if bin == "" {
		bin = "restic"
	}
	r := NewRunner(bin)
	r.RetryLock = 0
	// restic needs a cache directory; keep it inside the test
	r.Env = append(SafeEnv(), "RESTIC_CACHE_DIR="+t.TempDir())
	v, err := r.CheckVersion(context.Background())
	require.NoError(t, err, "restic %s", v)
	t.Logf("restic %s", v)
	return r
}

func writeTree(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	files := map[string]string{
		"docs/report.txt":       "quarterly report",
		"docs/notes/todo.md":    "- write tests",
		"docs/cache/big.bin":    strings.Repeat("x", 4096),
		"My Pictures/photo.jpg": "not really a jpeg",
	}
	for name, content := range files {
		path := filepath.Join(src, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return src
}

func TestIntegrationLocal(t *testing.T) {
	r := realRunner(t)
	repo := &Repo{Location: filepath.Join(t.TempDir(), "repo"), Password: "integration-password"}
	exercise(t, r, repo)
}

func TestIntegrationS3(t *testing.T) {
	location := os.Getenv("RESTIC_TEST_S3_LOCATION")
	if location == "" {
		t.Skip("RESTIC_TEST_S3_LOCATION is not set")
	}
	r := realRunner(t)
	repo := &Repo{
		// a fresh prefix per run so reruns start from an empty repository
		Location: location + "/" + time.Now().Format("20060102-150405.000000000"),
		Password: "integration-password",
		Env: map[string]string{
			"AWS_ACCESS_KEY_ID":     os.Getenv("AWS_ACCESS_KEY_ID"),
			"AWS_SECRET_ACCESS_KEY": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		},
	}
	exercise(t, r, repo)
}

func exercise(t *testing.T, r *Runner, repo *Repo) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	src := writeTree(t)

	exists, err := r.Exists(ctx, repo)
	require.NoError(t, err)
	assert.False(t, exists)
	require.NoError(t, r.Init(ctx, repo))
	exists, err = r.Exists(ctx, repo)
	require.NoError(t, err)
	assert.True(t, exists)

	backup := BackupOptions{
		Paths:    []string{filepath.Join(src, "docs"), filepath.Join(src, "My Pictures")},
		Excludes: []string{filepath.Join(src, "docs", "cache")},
		Tags:     []string{"resticgui:1"},
		Host:     "integration",
	}
	// This backup can finish before restic's first status tick, so live
	// progress is checked by TestIntegrationCancelReleasesLock instead.
	first, err := r.Backup(ctx, repo, backup, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, first.Summary.SnapshotID)
	assert.EqualValues(t, 3, first.Summary.FilesNew, "the excluded cache is not backed up")

	require.NoError(t, os.WriteFile(filepath.Join(src, "docs", "report.txt"), []byte("quarterly report v2"), 0o644))
	second, err := r.Backup(ctx, repo, backup, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 1, second.Summary.FilesChanged)

	snaps, err := r.Snapshots(ctx, repo, SnapshotFilter{Tags: []string{"resticgui:1"}, Host: "integration"})
	require.NoError(t, err)
	require.Len(t, snaps, 2)
	assert.Equal(t, second.Summary.SnapshotID, snaps[1].ID)

	nodes, err := r.Ls(ctx, repo, "latest", filepath.Join(src, "docs"))
	require.NoError(t, err)
	var names []string
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	assert.Contains(t, names, "report.txt")
	assert.NotContains(t, names, "cache")

	found, err := r.Find(ctx, repo, "latest", "todo.md")
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, 1, found[0].Hits)

	var dumped bytes.Buffer
	require.NoError(t, r.Dump(ctx, repo, "latest", filepath.Join(src, "docs", "report.txt"), "", &dumped))
	assert.Equal(t, "quarterly report v2", dumped.String())

	target := t.TempDir()
	restored, err := r.Restore(ctx, repo, RestoreOptions{SnapshotID: snaps[0].ID, Target: target}, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, restored.FilesRestored, uint64(3), "restic counts directories too")
	content, err := os.ReadFile(filepath.Join(target, src, "docs", "report.txt"))
	require.NoError(t, err)
	assert.Equal(t, "quarterly report", string(content), "the first snapshot has the first version")

	stats, err := r.Stats(ctx, repo, "raw-data", "")
	require.NoError(t, err)
	assert.Equal(t, 2, stats.SnapshotsCount)

	policy := ForgetOptions{Policy: Policy{KeepLast: 1}, Filter: SnapshotFilter{Tags: []string{"resticgui:1"}}, GroupBy: "host,tags"}
	policy.DryRun = true
	groups, err := r.Forget(ctx, repo, policy)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0].Remove, 1)
	snaps, err = r.Snapshots(ctx, repo, SnapshotFilter{})
	require.NoError(t, err)
	assert.Len(t, snaps, 2, "dry run removes nothing")

	policy.DryRun, policy.Prune = false, true
	_, err = r.Forget(ctx, repo, policy)
	require.NoError(t, err)
	snaps, err = r.Snapshots(ctx, repo, SnapshotFilter{})
	require.NoError(t, err)
	require.Len(t, snaps, 1)
	assert.Equal(t, second.Summary.SnapshotID, snaps[0].ID)

	check, err := r.Check(ctx, repo, CheckOptions{ReadDataSubset: "100%"})
	require.NoError(t, err)
	assert.Zero(t, check.NumErrors)

	require.NoError(t, r.Unlock(ctx, repo))

	wrong := *repo
	wrong.Password = "not-the-password"
	_, err = r.Snapshots(ctx, &wrong, SnapshotFilter{})
	assert.ErrorIs(t, err, ErrWrongPassword)
	_, err = r.Exists(ctx, &wrong)
	assert.ErrorIs(t, err, ErrWrongPassword)

	missing := *repo
	missing.Location = repo.Location + "-missing"
	_, err = r.Snapshots(ctx, &missing, SnapshotFilter{})
	assert.ErrorIs(t, err, ErrRepositoryNotFound)

	_, err = r.Backup(ctx, repo, BackupOptions{Paths: []string{filepath.Join(src, "does-not-exist")}}, nil)
	var rerr *Error
	require.True(t, errors.As(err, &rerr))
	assert.Contains(t, rerr.Message, "do not exist")

	// unknown IDs are reported, and one beginning with "-" is not taken for a flag
	err = r.ForgetSnapshots(ctx, repo, []string{"-h", "deadbeef"}, false)
	var notFound *SnapshotNotFoundError
	require.ErrorAs(t, err, &notFound)
	assert.Equal(t, []string{"-h", "deadbeef"}, notFound.IDs)
	require.NoError(t, r.ForgetSnapshots(ctx, repo, []string{snaps[0].ShortID}, true))
	snaps, err = r.Snapshots(ctx, repo, SnapshotFilter{})
	require.NoError(t, err)
	assert.Empty(t, snaps)
}

func TestIntegrationCancelReleasesLock(t *testing.T) {
	r := realRunner(t)
	ctx := context.Background()
	repo := &Repo{Location: filepath.Join(t.TempDir(), "repo"), Password: "integration-password"}
	require.NoError(t, r.Init(ctx, repo))

	// incompressible data, so the backup lasts long enough to be cancelled
	src := t.TempDir()
	f, err := os.Create(filepath.Join(src, "big.bin"))
	require.NoError(t, err)
	rng := rand.New(rand.NewPCG(1, 2))
	buf := make([]byte, 1<<20)
	for i := 0; i < 512; i++ {
		for j := range buf {
			buf[j] = byte(rng.Uint32())
		}
		_, err := f.Write(buf)
		require.NoError(t, err)
	}
	require.NoError(t, f.Close())

	// cancel from the progress callback, which also proves status messages
	// stream while the backup runs
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var progress int
	_, err = r.Backup(cctx, repo, BackupOptions{Paths: []string{src}}, func(s BackupStatus) {
		progress++
		if s.BytesDone > 0 {
			cancel()
		}
	})
	require.ErrorIs(t, err, context.Canceled)
	assert.Positive(t, progress)

	snaps, err := r.Snapshots(ctx, repo, SnapshotFilter{})
	require.NoError(t, err)
	assert.Empty(t, snaps, "no snapshot from a cancelled backup")
	// prune needs an exclusive lock: it fails with ErrLocked if the cancelled
	// backup left its lock behind
	require.NoError(t, r.Prune(ctx, repo))
}
