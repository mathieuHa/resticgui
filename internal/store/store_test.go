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
		Secrets: []byte("sealed"),
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
	assert.EqualValues(t, 6, version)

	// each down step works on its own
	_, err = migrator.Down(ctx)
	require.NoError(t, err)
	var subset int
	require.NoError(t, s.DB.QueryRowContext(ctx, `SELECT count(*) FROM pragma_table_info('schedules') WHERE name = 'read_data_subset'`).Scan(&subset))
	assert.Zero(t, subset)

	_, err = migrator.Down(ctx)
	require.NoError(t, err)
	var plans int
	require.NoError(t, s.DB.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name LIKE 'plan%'`).Scan(&plans))
	assert.Zero(t, plans)

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

	base := db.CreateRepositoryParams{Backend: "local", Location: "/x", Options: "{}", Secrets: []byte("s")}
	tests := map[string]func(p *db.CreateRepositoryParams){
		"duplicate name":  func(p *db.CreateRepositoryParams) { p.Name = "docs" },
		"empty name":      func(p *db.CreateRepositoryParams) { p.Name = "" },
		"unknown backend": func(p *db.CreateRepositoryParams) { p.Name = "a"; p.Backend = "ftp" },
		"empty location":  func(p *db.CreateRepositoryParams) { p.Name = "b"; p.Location = "" },
		"invalid options": func(p *db.CreateRepositoryParams) { p.Name = "c"; p.Options = "{not json" },
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
		Options: `{"env":{"AWS_DEFAULT_REGION":"eu-west-3"}}`,
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
		savePlan(t, s, id, "Documents")
		_, err := s.UpsertSchedule(ctx, db.UpsertScheduleParams{RepositoryID: id, Job: "check", Cron: "0 2 * * *", Enabled: true})
		require.NoError(t, err)
		_, err = s.CreateJob(ctx, db.CreateJobParams{RepositoryID: id, Kind: "check", Trigger: "manual", Params: "{}"})
		require.NoError(t, err)
	}

	n, err := s.DeleteRepository(ctx, r.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)

	for table, want := range map[string]int{"plans": 1, "plan_paths": 1, "plan_schedules": 1, "plan_retention": 1, "schedules": 1, "jobs": 1} {
		var count int
		require.NoError(t, s.DB.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count))
		assert.Equal(t, want, count, table)
	}
}

func TestForeignKeysAreEnforced(t *testing.T) {
	s := openTestStore(t)
	err := s.AddPlanPath(context.Background(), db.AddPlanPathParams{PlanID: 42, Path: "/data", Kind: "include"})
	assert.ErrorContains(t, err, "FOREIGN KEY")
}

func savePlan(t *testing.T, s *Store, repositoryID int64, name string) db.Plan {
	t.Helper()
	var plan db.Plan
	require.NoError(t, s.InTx(context.Background(), func(q *db.Queries) error {
		var err error
		plan, err = SavePlan(context.Background(), q, repositoryID, 0, Plan{
			Name: name, Include: []string{"/data"},
			Schedules: []db.PlanSchedule{{Job: "backup", Cron: "0 2 * * *", Enabled: true}},
			Retention: &db.PlanRetention{KeepLast: 3, Prune: true},
		})
		return err
	}))
	return plan
}

func TestPlans(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	r := createRepository(t, s, "docs")

	var plan db.Plan
	require.NoError(t, s.InTx(ctx, func(q *db.Queries) error {
		var err error
		plan, err = SavePlan(ctx, q, r.ID, 0, Plan{
			Name: "Home", Tags: []string{"laptop"},
			Options: PlanOptions{Compression: "max", OneFileSystem: true, LimitUpload: 512},
			Include: []string{"/home/me", "/etc", "/etc"}, Exclude: []string{"/home/me/.cache"},
			Schedules: []db.PlanSchedule{{Job: "backup", Cron: "@hourly", Enabled: true}, {Job: "forget", Cron: "0 4 * * 0", Enabled: false}},
			Retention: &db.PlanRetention{KeepDaily: 7, KeepWithin: "30d", Prune: true},
		})
		return err
	}))
	assert.Regexp(t, `^resticgui:\d+:[0-9a-f]{8}$`, plan.Tag)

	paths, err := s.ListPlanPaths(ctx, plan.ID)
	require.NoError(t, err)
	var got []string
	for _, p := range paths {
		got = append(got, p.Kind+":"+p.Path)
	}
	assert.Equal(t, []string{"exclude:/home/me/.cache", "include:/etc", "include:/home/me"}, got, "duplicates are dropped")
	opts, err := DecodePlanOptions(plan.Options)
	require.NoError(t, err)
	assert.Equal(t, PlanOptions{Compression: "max", OneFileSystem: true, LimitUpload: 512, ExcludeIfPresent: []string{}}, opts)
	tags, err := DecodeTags(plan.Tags)
	require.NoError(t, err)
	assert.Equal(t, []string{"laptop"}, tags)

	enabled, err := s.EnabledSchedules(ctx)
	require.NoError(t, err)
	assert.Equal(t, []Schedule{{RepositoryID: r.ID, PlanID: plan.ID, Job: "backup", Cron: "@hourly"}}, enabled)

	// saving replaces everything but the tag
	require.NoError(t, s.InTx(ctx, func(q *db.Queries) error {
		updated, err := SavePlan(ctx, q, r.ID, plan.ID, Plan{Name: "Home folder", Include: []string{"/home/me"}})
		assert.Equal(t, plan.Tag, updated.Tag)
		return err
	}))
	schedules, err := s.ListPlanSchedules(ctx, plan.ID)
	require.NoError(t, err)
	assert.Empty(t, schedules)
	_, err = s.GetPlanRetention(ctx, plan.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	err = s.InTx(ctx, func(q *db.Queries) error {
		_, err := SavePlan(ctx, q, r.ID, 0, Plan{Name: "Home folder"})
		return err
	})
	assert.Error(t, err, "names are unique within a repository")
	other := createRepository(t, s, "other")
	savePlan(t, s, other.ID, "Home folder")

	err = s.InTx(ctx, func(q *db.Queries) error {
		_, err := SavePlan(ctx, q, r.ID, plan.ID, Plan{Name: "x", Retention: &db.PlanRetention{KeepLast: -1}})
		return err
	})
	assert.Error(t, err, "negative counts are rejected")
	err = s.InTx(ctx, func(q *db.Queries) error {
		_, err := SavePlan(ctx, q, r.ID, plan.ID, Plan{Name: "x", Schedules: []db.PlanSchedule{{Job: "check", Cron: "@daily"}}})
		return err
	})
	assert.Error(t, err, "a check is scheduled on the repository")

	n, err := s.DeletePlan(ctx, plan.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)
	paths, err = s.ListPlanPaths(ctx, plan.ID)
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
			Name: "rolled-back", Backend: "local", Location: "/x", Options: "{}", Secrets: []byte("s"),
		})
		require.NoError(t, err)
		return boom
	})
	assert.ErrorIs(t, err, boom)
	_, err = s.GetRepositoryByName(ctx, "rolled-back")
	assert.ErrorIs(t, err, sql.ErrNoRows)

	require.NoError(t, s.InTx(ctx, func(q *db.Queries) error {
		_, err := q.CreateRepository(ctx, db.CreateRepositoryParams{
			Name: "committed", Backend: "local", Location: "/x", Options: "{}", Secrets: []byte("s"),
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
		Name: "s3", Backend: "s3", Location: "s3:s3.amazonaws.com/b", Options: "{}", Secrets: sealed,
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

// A database of phase 8, with a repository set up the way it was then,
// upgrades to one plan holding its settings, and back.
func TestMigrationTurnsRepositorySettingsIntoAPlan(t *testing.T) {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", dsn(filepath.Join(t.TempDir(), "resticgui.db")))
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { conn.Close() })
	migrator, err := newMigrator(conn)
	require.NoError(t, err)
	_, err = migrator.UpTo(ctx, 3)
	require.NoError(t, err)

	for _, stmt := range []string{
		`INSERT INTO repositories (id, name, backend, location, secrets, host, compression, exclude_larger_than)
		 VALUES (7, 'photos', 'local', '/srv/photos', x'00', 'nas', 'max', '2G'), (8, 'empty', 'local', '/srv/empty', x'00', '', 'auto', '')`,
		`INSERT INTO paths (repository_id, path, kind) VALUES (7, '/photos', 'include'), (7, '*.tmp', 'exclude')`,
		`INSERT INTO schedules (repository_id, job, cron, enabled) VALUES (7, 'backup', '30 2 * * *', 1), (7, 'forget', '0 4 * * 0', 0), (7, 'check', '@monthly', 1)`,
		`INSERT INTO retention (repository_id, keep_daily, keep_within, prune) VALUES (7, 7, '1y', 0)`,
		`INSERT INTO jobs (repository_id, kind, trigger, status) VALUES (7, 'backup', 'schedule', 'succeeded'), (7, 'check', 'manual', 'succeeded')`,
	} {
		_, err := conn.ExecContext(ctx, stmt)
		require.NoError(t, err, stmt)
	}

	_, err = migrator.Up(ctx)
	require.NoError(t, err)
	q := db.New(conn)
	plans, err := q.ListPlans(ctx, 7)
	require.NoError(t, err)
	require.Len(t, plans, 1)
	plan := plans[0]
	assert.Equal(t, "Default", plan.Name)
	assert.Equal(t, "resticgui:7", plan.Tag, "the tag of the snapshots made so far")
	opts, err := DecodePlanOptions(plan.Options)
	require.NoError(t, err)
	assert.Equal(t, "max", opts.Compression)
	assert.Equal(t, "2G", opts.ExcludeLargerThan)
	paths, err := q.ListPlanPaths(ctx, plan.ID)
	require.NoError(t, err)
	assert.Len(t, paths, 2)
	schedules, err := q.ListPlanSchedules(ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.PlanSchedule{
		{PlanID: plan.ID, Job: "backup", Cron: "30 2 * * *", Enabled: true},
		{PlanID: plan.ID, Job: "forget", Cron: "0 4 * * 0", Enabled: false},
	}, schedules)
	repoSchedules, err := q.ListSchedules(ctx, 7)
	require.NoError(t, err)
	require.Len(t, repoSchedules, 1)
	assert.Equal(t, "check", repoSchedules[0].Job, "the check stays on the repository")
	ret, err := q.GetPlanRetention(ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, db.PlanRetention{PlanID: plan.ID, KeepDaily: 7, KeepWithin: "1y"}, ret)
	history, err := q.ListJobsForRepository(ctx, db.ListJobsForRepositoryParams{RepositoryID: 7, Limit: 10})
	require.NoError(t, err)
	require.Len(t, history, 2)
	assert.Nil(t, history[0].PlanID, "a check belongs to no plan")
	require.NotNil(t, history[1].PlanID)
	assert.Equal(t, plan.ID, *history[1].PlanID)
	empty, err := q.ListPlans(ctx, 8)
	require.NoError(t, err)
	assert.Len(t, empty, 1, "every repository gets its plan")

	// a second plan is lost going down; the first one's settings come back
	require.NoError(t, (&Store{Queries: q, DB: conn}).InTx(ctx, func(tx *db.Queries) error {
		_, err := SavePlan(ctx, tx, 7, 0, Plan{Name: "Videos", Include: []string{"/videos"}})
		return err
	}))
	_, err = migrator.DownTo(ctx, 4)
	require.NoError(t, err)
	var compression, larger string
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT compression, exclude_larger_than FROM repositories WHERE id = 7`).Scan(&compression, &larger))
	assert.Equal(t, []string{"max", "2G"}, []string{compression, larger})
	var restored []string
	rows, err := conn.QueryContext(ctx, `SELECT kind || ':' || path FROM paths WHERE repository_id = 7 ORDER BY kind, path`)
	require.NoError(t, err)
	for rows.Next() {
		var p string
		require.NoError(t, rows.Scan(&p))
		restored = append(restored, p)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, []string{"exclude:*.tmp", "include:/photos"}, restored)
	var count int
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT count(*) FROM schedules WHERE repository_id = 7`).Scan(&count))
	assert.Equal(t, 3, count)
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT keep_daily FROM retention WHERE repository_id = 7`).Scan(&count))
	assert.Equal(t, 7, count)
}

func TestMigrationAddsTheMaintenanceJobs(t *testing.T) {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", dsn(filepath.Join(t.TempDir(), "resticgui.db")))
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { conn.Close() })
	migrator, err := newMigrator(conn)
	require.NoError(t, err)
	_, err = migrator.UpTo(ctx, 5)
	require.NoError(t, err)
	for _, stmt := range []string{
		`INSERT INTO repositories (id, name, backend, location, secrets) VALUES (7, 'photos', 'local', '/srv/photos', x'00')`,
		`INSERT INTO schedules (repository_id, job, cron) VALUES (7, 'check', '@monthly')`,
		`INSERT INTO jobs (repository_id, plan_id, kind, trigger, status, summary, params) VALUES (7, 7, 'backup', 'schedule', 'succeeded', '{"files_new":1}', '{}')`,
	} {
		_, err := conn.ExecContext(ctx, stmt)
		require.NoError(t, err, stmt)
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO jobs (repository_id, kind, trigger) VALUES (7, 'repair_index', 'manual')`)
	require.Error(t, err, "not a kind yet")

	_, err = migrator.Up(ctx)
	require.NoError(t, err)
	q := db.New(conn)
	history, err := q.ListJobsForRepository(ctx, db.ListJobsForRepositoryParams{RepositoryID: 7, Limit: 10})
	require.NoError(t, err)
	require.Len(t, history, 1, "the history is kept")
	assert.Equal(t, "backup", history[0].Kind)
	assert.JSONEq(t, `{"files_new":1}`, history[0].Summary)
	require.NotNil(t, history[0].PlanID)
	for _, kind := range []string{"repair_index", "repair_snapshots", "migrate", "change_password"} {
		_, err := q.CreateJob(ctx, db.CreateJobParams{RepositoryID: 7, Kind: kind, Trigger: "manual", Params: "{}"})
		require.NoError(t, err, kind)
	}
	schedules, err := q.ListSchedules(ctx, 7)
	require.NoError(t, err)
	assert.Equal(t, "", schedules[0].ReadDataSubset)

	// going down drops the runs of the new kinds only
	_, err = migrator.Down(ctx)
	require.NoError(t, err)
	var count int
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT count(*) FROM jobs`).Scan(&count))
	assert.Equal(t, 1, count)
	// the cascade from repositories still works on the table made again
	_, err = conn.ExecContext(ctx, `DELETE FROM repositories WHERE id = 7`)
	require.NoError(t, err)
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT count(*) FROM jobs`).Scan(&count))
	assert.Zero(t, count)
}
