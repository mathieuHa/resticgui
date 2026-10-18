package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(context.Background(), filepath.Join(t.TempDir(), "resticgui.db"))
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func testBox(t *testing.T) *secrets.Box {
	t.Helper()
	box, err := secrets.New(bytes.Repeat([]byte{3}, secrets.KeySize))
	require.NoError(t, err)
	return box
}

func createRepository(t *testing.T, s *Store, name string) db.Repository {
	t.Helper()
	r, err := s.CreateRepository(context.Background(), db.CreateRepositoryParams{
		Name: name, Backend: "local", Location: "/srv/restic/" + name, Options: "{}",
		Secrets: []byte("sealed"), Compression: "auto",
	})
	require.NoError(t, err)
	return r
}

func TestMigrationsUpDownUp(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	migrator, err := newMigrator(s.DB)
	require.NoError(t, err)

	version, err := migrator.GetDBVersion(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 4, version)

	// each down step works on its own
	_, err = migrator.Down(ctx)
	require.NoError(t, err)
	version, err = migrator.GetDBVersion(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 3, version)
	var sessions int
	require.NoError(t, s.DB.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'sessions'`).Scan(&sessions))
	assert.Zero(t, sessions)

	_, err = migrator.Down(ctx)
	require.NoError(t, err)
	version, err = migrator.GetDBVersion(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 2, version)
	var tables int
	require.NoError(t, s.DB.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN ('notifiers', 'healthchecks', 'hooks', 'repository_stats')`).Scan(&tables))
	assert.Zero(t, tables)

	_, err = migrator.DownTo(ctx, 0)
	require.NoError(t, err)
	require.NoError(t, s.DB.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN ('repositories', 'paths', 'schedules', 'retention', 'jobs')`).Scan(&tables))
	assert.Zero(t, tables)

	_, err = migrator.Up(ctx)
	require.NoError(t, err)
	createRepository(t, s, "after-reapply")
}

func TestOpenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resticgui.db")
	s, err := Open(context.Background(), path)
	require.NoError(t, err)
	createRepository(t, s, "kept")
	require.NoError(t, s.Close())

	s, err = Open(context.Background(), path)
	require.NoError(t, err)
	defer s.Close()
	r, err := s.GetRepositoryByName(context.Background(), "kept")
	require.NoError(t, err)
	assert.Equal(t, "/srv/restic/kept", r.Location)
}

func TestRepositoryConstraints(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	createRepository(t, s, "docs")

	base := db.CreateRepositoryParams{Backend: "local", Location: "/x", Options: "{}", Secrets: []byte("s"), Compression: "auto"}
	tests := map[string]func(p *db.CreateRepositoryParams){
		"duplicate name":      func(p *db.CreateRepositoryParams) { p.Name = "docs" },
		"empty name":          func(p *db.CreateRepositoryParams) { p.Name = "" },
		"unknown backend":     func(p *db.CreateRepositoryParams) { p.Name = "a"; p.Backend = "ftp" },
		"empty location":      func(p *db.CreateRepositoryParams) { p.Name = "b"; p.Location = "" },
		"invalid options":     func(p *db.CreateRepositoryParams) { p.Name = "c"; p.Options = "{not json" },
		"unknown compression": func(p *db.CreateRepositoryParams) { p.Name = "d"; p.Compression = "zstd" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			p := base
			mutate(&p)
			_, err := s.CreateRepository(ctx, p)
			assert.Error(t, err)
		})
	}
}

func TestUpdateRepository(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")

	updated, err := s.UpdateRepository(ctx, db.UpdateRepositoryParams{
		ID: r.ID, Name: "documents", Backend: "s3", Location: "s3:s3.amazonaws.com/docs",
		Options: `{"env":{"AWS_DEFAULT_REGION":"eu-west-3"}}`, Compression: "max",
	})
	require.NoError(t, err)
	assert.Equal(t, "documents", updated.Name)
	assert.Equal(t, r.Secrets, updated.Secrets, "secrets are only changed by UpdateRepositorySecrets")

	n, err := s.UpdateRepositorySecrets(ctx, db.UpdateRepositorySecretsParams{ID: r.ID, Secrets: []byte("new")})
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)

	require.NoError(t, s.MarkRepositoryInitialized(ctx, r.ID))
	got, err := s.GetRepository(ctx, r.ID)
	require.NoError(t, err)
	assert.NotNil(t, got.InitializedAt)
	assert.Equal(t, []byte("new"), got.Secrets)
}

func TestDeleteRepositoryCascades(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")
	other := createRepository(t, s, "other")

	for _, id := range []int64{r.ID, other.ID} {
		_, err := s.AddPath(ctx, db.AddPathParams{RepositoryID: id, Path: "/data", Kind: "include"})
		require.NoError(t, err)
		_, err = s.UpsertSchedule(ctx, db.UpsertScheduleParams{RepositoryID: id, Job: "backup", Cron: "0 2 * * *", Enabled: true})
		require.NoError(t, err)
		_, err = s.UpsertRetention(ctx, db.UpsertRetentionParams{RepositoryID: id, KeepLast: 3, Prune: true})
		require.NoError(t, err)
		_, err = s.CreateJob(ctx, db.CreateJobParams{RepositoryID: id, Kind: "backup", Trigger: "manual", Params: "{}"})
		require.NoError(t, err)
	}

	n, err := s.DeleteRepository(ctx, r.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)

	for table, want := range map[string]int{"paths": 1, "schedules": 1, "retention": 1, "jobs": 1} {
		var count int
		require.NoError(t, s.DB.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count))
		assert.Equal(t, want, count, table)
	}
}

func TestForeignKeysAreEnforced(t *testing.T) {
	s := openTestStore(t)
	_, err := s.AddPath(context.Background(), db.AddPathParams{RepositoryID: 42, Path: "/data", Kind: "include"})
	assert.ErrorContains(t, err, "FOREIGN KEY")
}

func TestPaths(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")

	for _, p := range []db.AddPathParams{
		{RepositoryID: r.ID, Path: "/home/me", Kind: "include"},
		{RepositoryID: r.ID, Path: "/home/me/.cache", Kind: "exclude"},
		{RepositoryID: r.ID, Path: "/etc", Kind: "include"},
	} {
		_, err := s.AddPath(ctx, p)
		require.NoError(t, err)
	}
	_, err := s.AddPath(ctx, db.AddPathParams{RepositoryID: r.ID, Path: "/etc", Kind: "include"})
	assert.Error(t, err, "duplicate path")
	_, err = s.AddPath(ctx, db.AddPathParams{RepositoryID: r.ID, Path: "/opt", Kind: "maybe"})
	assert.Error(t, err, "unknown kind")

	paths, err := s.ListPaths(ctx, r.ID)
	require.NoError(t, err)
	var got []string
	for _, p := range paths {
		got = append(got, p.Kind+":"+p.Path)
	}
	assert.Equal(t, []string{"exclude:/home/me/.cache", "include:/etc", "include:/home/me"}, got)

	require.NoError(t, s.DeletePaths(ctx, r.ID))
	paths, err = s.ListPaths(ctx, r.ID)
	require.NoError(t, err)
	assert.Empty(t, paths)
}

func TestSchedules(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")

	first, err := s.UpsertSchedule(ctx, db.UpsertScheduleParams{RepositoryID: r.ID, Job: "backup", Cron: "0 2 * * *", Enabled: true})
	require.NoError(t, err)
	second, err := s.UpsertSchedule(ctx, db.UpsertScheduleParams{RepositoryID: r.ID, Job: "backup", Cron: "@every 6h", Enabled: false})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID, "upsert keeps one schedule per job")
	assert.Equal(t, "@every 6h", second.Cron)

	_, err = s.UpsertSchedule(ctx, db.UpsertScheduleParams{RepositoryID: r.ID, Job: "forget", Cron: "0 4 * * 0", Enabled: true})
	require.NoError(t, err)
	_, err = s.UpsertSchedule(ctx, db.UpsertScheduleParams{RepositoryID: r.ID, Job: "restore", Cron: "0 4 * * 0", Enabled: true})
	assert.Error(t, err, "restore is not a schedulable job")

	all, err := s.ListSchedules(ctx, r.ID)
	require.NoError(t, err)
	assert.Len(t, all, 2)
	enabled, err := s.ListEnabledSchedules(ctx)
	require.NoError(t, err)
	require.Len(t, enabled, 1)
	assert.Equal(t, "forget", enabled[0].Job)

	n, err := s.DeleteSchedule(ctx, db.DeleteScheduleParams{RepositoryID: r.ID, Job: "forget"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)
}

func TestRetention(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")

	_, err := s.GetRetention(ctx, r.ID)
	assert.True(t, errors.Is(err, sql.ErrNoRows))

	_, err = s.UpsertRetention(ctx, db.UpsertRetentionParams{RepositoryID: r.ID, KeepDaily: 7, Prune: true})
	require.NoError(t, err)
	got, err := s.UpsertRetention(ctx, db.UpsertRetentionParams{RepositoryID: r.ID, KeepHourly: 24, KeepYearly: 2, KeepWithin: "30d"})
	require.NoError(t, err)
	assert.Equal(t, db.Retention{RepositoryID: r.ID, KeepHourly: 24, KeepYearly: 2, KeepWithin: "30d"}, got)

	_, err = s.UpsertRetention(ctx, db.UpsertRetentionParams{RepositoryID: r.ID, KeepLast: -1})
	assert.Error(t, err, "negative counts are rejected")
}

func TestJobLifecycle(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")

	job, err := s.CreateJob(ctx, db.CreateJobParams{RepositoryID: r.ID, Kind: "backup", Trigger: "schedule", Params: "{}"})
	require.NoError(t, err)
	assert.Equal(t, "queued", job.Status)
	assert.Nil(t, job.StartedAt)

	n, err := s.StartJob(ctx, job.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)
	n, err = s.StartJob(ctx, job.ID)
	require.NoError(t, err)
	assert.Zero(t, n, "a running job cannot be started again")

	exitCode := int64(3)
	finish := db.FinishJobParams{
		ID: job.ID, Status: "failed", SnapshotID: "4f1c2a", Summary: `{"files_new":2}`,
		Error: "some files could not be read", ExitCode: &exitCode, LogExcerpt: "permission denied",
	}
	n, err = s.FinishJob(ctx, finish)
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)
	n, err = s.FinishJob(ctx, finish)
	require.NoError(t, err)
	assert.Zero(t, n, "a finished job is immutable")

	got, err := s.GetJob(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, "failed", got.Status)
	assert.NotNil(t, got.StartedAt)
	assert.NotNil(t, got.FinishedAt)
	require.NotNil(t, got.ExitCode)
	assert.EqualValues(t, 3, *got.ExitCode)

	finish.Summary = "not json"
	other, err := s.CreateJob(ctx, db.CreateJobParams{RepositoryID: r.ID, Kind: "backup", Trigger: "manual", Params: "{}"})
	require.NoError(t, err)
	finish.ID = other.ID
	_, err = s.FinishJob(ctx, finish)
	assert.Error(t, err, "summary must be JSON")
}

func TestListJobsAndInterrupt(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")

	var ids []int64
	for i := 0; i < 5; i++ {
		job, err := s.CreateJob(ctx, db.CreateJobParams{RepositoryID: r.ID, Kind: "backup", Trigger: "manual", Params: "{}"})
		require.NoError(t, err)
		ids = append(ids, job.ID)
	}
	_, err := s.StartJob(ctx, ids[4])
	require.NoError(t, err)
	_, err = s.FinishJob(ctx, db.FinishJobParams{ID: ids[0], Status: "succeeded", Summary: "{}"})
	require.NoError(t, err)

	page, err := s.ListJobsForRepository(ctx, db.ListJobsForRepositoryParams{RepositoryID: r.ID, Limit: 2, Offset: 1})
	require.NoError(t, err)
	require.Len(t, page, 2)
	assert.Equal(t, []int64{ids[3], ids[2]}, []int64{page[0].ID, page[1].ID}, "newest first")

	n, err := s.MarkUnfinishedJobsInterrupted(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 4, n)
	first, err := s.GetJob(ctx, ids[0])
	require.NoError(t, err)
	assert.Equal(t, "succeeded", first.Status, "finished jobs are left alone")
}

func TestInTxRollsBack(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	boom := errors.New("boom")

	err := s.InTx(ctx, func(q *db.Queries) error {
		_, err := q.CreateRepository(ctx, db.CreateRepositoryParams{
			Name: "rolled-back", Backend: "local", Location: "/x", Options: "{}", Secrets: []byte("s"), Compression: "auto",
		})
		require.NoError(t, err)
		return boom
	})
	assert.ErrorIs(t, err, boom)
	_, err = s.GetRepositoryByName(ctx, "rolled-back")
	assert.ErrorIs(t, err, sql.ErrNoRows)

	require.NoError(t, s.InTx(ctx, func(q *db.Queries) error {
		_, err := q.CreateRepository(ctx, db.CreateRepositoryParams{
			Name: "committed", Backend: "local", Location: "/x", Options: "{}", Secrets: []byte("s"), Compression: "auto",
		})
		return err
	}))
	_, err = s.GetRepositoryByName(ctx, "committed")
	assert.NoError(t, err)
}

func TestSecretsAreStoredEncrypted(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	box := testBox(t)

	in := RepositorySecrets{Password: "correct horse battery staple", Env: map[string]string{"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI"}}
	sealed, err := SealSecrets(box, in)
	require.NoError(t, err)
	r, err := s.CreateRepository(ctx, db.CreateRepositoryParams{
		Name: "s3", Backend: "s3", Location: "s3:s3.amazonaws.com/b", Options: "{}", Secrets: sealed, Compression: "auto",
	})
	require.NoError(t, err)

	var raw []byte
	require.NoError(t, s.DB.QueryRowContext(ctx, "SELECT secrets FROM repositories WHERE id = ?", r.ID).Scan(&raw))
	assert.NotContains(t, string(raw), in.Password)
	assert.NotContains(t, string(raw), "wJalrXUtnFEMI")

	out, err := OpenSecrets(box, r.Secrets)
	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestOptionsRoundTrip(t *testing.T) {
	in := RepositoryOptions{Env: map[string]string{"AWS_DEFAULT_REGION": "eu-west-3"}, CACert: "/etc/ca.pem", Extended: map[string]string{"s3.storage-class": "STANDARD_IA"}}
	encoded, err := EncodeOptions(in)
	require.NoError(t, err)
	out, err := DecodeOptions(encoded)
	require.NoError(t, err)
	assert.Equal(t, in, out)

	empty, err := EncodeOptions(RepositoryOptions{})
	require.NoError(t, err)
	assert.Equal(t, "{}", empty)
}

func TestRepositoryMonitoringGoesWithTheRepository(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "photos")
	require.NoError(t, s.UpsertHealthcheck(ctx, db.UpsertHealthcheckParams{RepositoryID: r.ID, Url: []byte("sealed"), Host: "hc-ping.com"}))
	_, err := s.CreateHook(ctx, db.CreateHookParams{RepositoryID: r.ID, Event: "before", Kind: "command", Command: "true", TimeoutSeconds: 60})
	require.NoError(t, err)
	require.NoError(t, s.UpsertRepositoryStats(ctx, db.UpsertRepositoryStatsParams{RepositoryID: r.ID, Snapshots: 2, TotalSize: 10}))
	require.NoError(t, s.UpsertRepositoryStats(ctx, db.UpsertRepositoryStatsParams{RepositoryID: r.ID, Snapshots: 3, TotalSize: 20}))
	stats, err := s.ListRepositoryStats(ctx)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.EqualValues(t, 3, stats[0].Snapshots)

	_, err = s.CreateHook(ctx, db.CreateHookParams{RepositoryID: r.ID, Event: "after", Kind: "command", Command: "true", TimeoutSeconds: 60})
	assert.Error(t, err, "the event is checked")

	_, err = s.DeleteRepository(ctx, r.ID)
	require.NoError(t, err)
	hooks, err := s.ListAllHooks(ctx)
	require.NoError(t, err)
	assert.Empty(t, hooks)
	checks, err := s.ListHealthchecks(ctx)
	require.NoError(t, err)
	assert.Empty(t, checks)
	stats, err = s.ListRepositoryStats(ctx)
	require.NoError(t, err)
	assert.Empty(t, stats)
}
