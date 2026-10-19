package legacy

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

const fixture = "testdata/legacy-v0.1.10.db"

func testBox(t *testing.T) *secrets.Box {
	t.Helper()
	box, err := secrets.New(bytes.Repeat([]byte{9}, secrets.KeySize))
	require.NoError(t, err)
	return box
}

func openStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "new.db"))
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func importFixture(t *testing.T) (*store.Store, Report) {
	t.Helper()
	s := openStore(t)
	report, err := Import(context.Background(), fixture, s, testBox(t))
	require.NoError(t, err)
	return s, report
}

type imported struct {
	repo      db.Repository
	options   store.RepositoryOptions
	secrets   store.RepositorySecrets
	plan      store.PlanOptions
	paths     []string
	schedules map[string]db.PlanSchedule
	retention *db.PlanRetention
}

func load(t *testing.T, s *store.Store, name string) imported {
	t.Helper()
	ctx := context.Background()
	r, err := s.GetRepositoryByName(ctx, name)
	require.NoError(t, err)
	out := imported{repo: r, schedules: map[string]db.PlanSchedule{}}
	out.options, err = store.DecodeOptions(r.Options)
	require.NoError(t, err)
	out.secrets, err = store.OpenSecrets(testBox(t), r.Secrets)
	require.NoError(t, err)
	plans, err := s.ListPlans(ctx, r.ID)
	require.NoError(t, err)
	require.Len(t, plans, 1, "one plan per v0.1 repository")
	plan := plans[0]
	out.plan, err = store.DecodePlanOptions(plan.Options)
	require.NoError(t, err)
	paths, err := s.ListPlanPaths(ctx, plan.ID)
	require.NoError(t, err)
	for _, p := range paths {
		out.paths = append(out.paths, p.Kind+":"+p.Path)
	}
	schedules, err := s.ListPlanSchedules(ctx, plan.ID)
	require.NoError(t, err)
	for _, sc := range schedules {
		out.schedules[sc.Job] = sc
	}
	if ret, err := s.GetPlanRetention(ctx, plan.ID); err == nil {
		ret.PlanID = 0
		out.retention = &ret
	}
	return out
}

func TestImportFixture(t *testing.T) {
	s, report := importFixture(t)

	assert.Equal(t, []string{"local-docs", "aws-photos", "b2-archive", "minio-lab"}, report.Imported)
	assert.Equal(t, []string{"unconfigured: no storage backend configured"}, report.Skipped)
	require.Len(t, report.Warnings, 1)
	assert.Contains(t, report.Warnings[0], "minio-lab: forget schedule not imported")
	assert.Equal(t, 2, report.LogsIgnored)

	t.Run("local", func(t *testing.T) {
		r := load(t, s, "local-docs")
		assert.Equal(t, "local", r.repo.Backend)
		assert.Equal(t, "/srv/restic/docs", r.repo.Location)
		assert.Equal(t, "2G", r.plan.ExcludeLargerThan)
		assert.Nil(t, r.repo.InitializedAt, "v0.1's initiated flag is not trusted")
		assert.Equal(t, store.RepositorySecrets{Password: "docs-password"}, r.secrets)
		assert.Equal(t, []string{"exclude:/home/me/docs/tmp", "include:/home/me/docs"}, r.paths)
		assert.Equal(t, "30 2 * * 1,5", r.schedules["backup"].Cron)
		assert.True(t, r.schedules["backup"].Enabled)
		assert.Equal(t, "0 4 * * 0", r.schedules["forget"].Cron)
		assert.True(t, r.schedules["forget"].Enabled)
		require.NotNil(t, r.retention)
		assert.Equal(t, db.PlanRetention{KeepLast: 5, KeepDaily: 7, KeepWeekly: 4, KeepMonthly: 6, Prune: true}, *r.retention)
	})

	t.Run("aws s3", func(t *testing.T) {
		r := load(t, s, "aws-photos")
		assert.Equal(t, "s3", r.repo.Backend)
		assert.Equal(t, "s3:s3.amazonaws.com/my-photos", r.repo.Location)
		assert.Equal(t, map[string]string{"AWS_DEFAULT_REGION": "eu-west-3"}, r.options.Env)
		assert.Equal(t, store.RepositorySecrets{
			Password: "photos-password",
			Env:      map[string]string{"AWS_ACCESS_KEY_ID": "AKIAEXAMPLE", "AWS_SECRET_ACCESS_KEY": "aws-secret-key"},
		}, r.secrets)
		assert.Equal(t, []string{"include:/home/me/pictures"}, r.paths, "falls back to abs_path when rel_path is empty")
		assert.Equal(t, "@every 1h30m0s", r.schedules["backup"].Cron)
		assert.False(t, r.schedules["backup"].Enabled, "inactive repository keeps its schedule disabled")
		require.NotNil(t, r.retention)
		assert.Equal(t, db.PlanRetention{KeepLast: 10, KeepHourly: 1, KeepDaily: 1, KeepWeekly: 1, KeepMonthly: 1, Prune: true}, *r.retention)
	})

	t.Run("b2", func(t *testing.T) {
		r := load(t, s, "b2-archive")
		assert.Equal(t, "b2", r.repo.Backend)
		assert.Equal(t, "b2:archive:laptop", r.repo.Location)
		assert.Equal(t, map[string]string{"B2_ACCOUNT_ID": "b2-account", "B2_ACCOUNT_KEY": "b2-secret-key"}, r.secrets.Env)
		assert.Equal(t, "15 1 * * *", r.schedules["backup"].Cron)
		assert.Nil(t, r.retention)
		assert.Empty(t, r.paths)
	})

	t.Run("minio", func(t *testing.T) {
		r := load(t, s, "minio-lab")
		assert.Equal(t, "s3", r.repo.Backend)
		assert.Equal(t, "s3:https://10.0.0.5:9000/lab", r.repo.Location)
		assert.Equal(t, "/etc/ssl/minio-ca.pem", r.options.CACert)
		assert.Equal(t, map[string]string{"AWS_ACCESS_KEY_ID": "minio-access", "AWS_SECRET_ACCESS_KEY": "minio-secret-key"}, r.secrets.Env)
		assert.Equal(t, "0 3 1 * *", r.schedules["backup"].Cron, "monthly, which v0.1 could not schedule")
		_, hasForget := r.schedules["forget"]
		assert.False(t, hasForget, "invalid cron spec is reported, not imported")
	})
}

func TestImportNeverStoresSecretsInClear(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.db")
	s, err := store.Open(context.Background(), path)
	require.NoError(t, err)
	_, err = Import(context.Background(), fixture, s, testBox(t))
	require.NoError(t, err)
	require.NoError(t, s.Close())

	for _, name := range []string{"new.db", "new.db-wal"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			continue
		}
		require.NoError(t, err)
		for _, secret := range []string{"docs-password", "photos-password", "aws-secret-key", "b2-secret-key", "minio-secret-key", "lab-password"} {
			assert.NotContains(t, string(raw), secret, "%s found in %s", secret, name)
		}
	}
}

func TestImportIsAllOrNothing(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	_, err := s.CreateRepository(ctx, db.CreateRepositoryParams{
		Name: "b2-archive", Backend: "local", Location: "/already/here", Options: "{}", Secrets: []byte("x"),
	})
	require.NoError(t, err)

	_, err = Import(ctx, fixture, s, testBox(t))
	require.ErrorContains(t, err, `repository "b2-archive"`)

	repos, err := s.ListRepositories(ctx)
	require.NoError(t, err)
	require.Len(t, repos, 1, "nothing from the failed import is kept")
}

func TestImportRejectsNonLegacyDatabase(t *testing.T) {
	other := filepath.Join(t.TempDir(), "other.db")
	s, err := store.Open(context.Background(), other)
	require.NoError(t, err)
	require.NoError(t, s.Close())

	_, err = Import(context.Background(), other, openStore(t), testBox(t))
	assert.ErrorContains(t, err, "is this a ResticGUI v0.1 database?")
}

func TestImportLeavesSourceUntouched(t *testing.T) {
	before, err := os.ReadFile(fixture)
	require.NoError(t, err)
	importFixture(t)
	after, err := os.ReadFile(fixture)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(before, after))
}

func TestLegacyScheduleCron(t *testing.T) {
	weekdays := func(days ...int) (w [7]bool) {
		for _, d := range days {
			w[d] = true
		}
		return w
	}
	tests := []struct {
		name    string
		in      legacySchedule
		want    string
		wantErr string
	}{
		{"cron", legacySchedule{choice: "cron", cronSyntax: " */30 * * * * "}, "*/30 * * * *", ""},
		{"cron descriptor", legacySchedule{choice: "cron", cronSyntax: "@daily"}, "@daily", ""},
		{"invalid cron", legacySchedule{choice: "cron", cronSyntax: "every sunday"}, "", "invalid cron spec"},
		{"empty cron", legacySchedule{choice: "cron"}, "", "invalid cron spec"},
		{"interval", legacySchedule{choice: "interval", interval: "6h"}, "@every 6h0m0s", ""},
		{"interval as the UI placeholder", legacySchedule{choice: "interval", interval: "1h30"}, "@every 1h30m0s", ""},
		{"invalid interval", legacySchedule{choice: "interval", interval: "soon"}, "", "invalid interval"},
		{"weekly", legacySchedule{choice: "weekly", hour: "22:05", weekdays: weekdays(0, 3)}, "5 22 * * 3,0", ""},
		{"weekly without days", legacySchedule{choice: "weekly", hour: "22:05"}, "", "without any day"},
		{"weekly bad hour", legacySchedule{choice: "weekly", hour: "25:00", weekdays: weekdays(1)}, "", "invalid hour"},
		{"monthly", legacySchedule{choice: "monthly", everyMonth: true, day: "15", hour: "01:00"}, "0 1 15 * *", ""},
		{"regular daily", legacySchedule{choice: "regular", everyDay: true, hour: "06:45"}, "45 6 * * *", ""},
		{"regular weekly", legacySchedule{choice: "regular", everyWeek: true, day: "2", hour: "06:45"}, "45 6 * * 2", ""},
		{"monthly without period", legacySchedule{choice: "monthly", hour: "01:00"}, "", "without a period"},
		{"unknown", legacySchedule{choice: "yearly"}, "", "unknown schedule type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.cron()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMain_CLI(t *testing.T) {
	t.Setenv(secrets.KeyEnv, "")
	os.Unsetenv(secrets.KeyEnv)
	dir := t.TempDir()
	to := filepath.Join(dir, "resticgui-new.db")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"-from", fixture, "-to", to}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	assert.Contains(t, stdout.String(), "Imported 4 repositories")
	assert.Contains(t, stdout.String(), "skipped unconfigured: no storage backend configured")
	assert.FileExists(t, filepath.Join(dir, "resticgui.key"))

	t.Run("refuses an existing target", func(t *testing.T) {
		stdout.Reset()
		stderr.Reset()
		assert.Equal(t, 1, Main([]string{"-from", fixture, "-to", to}, &stdout, &stderr))
		assert.Contains(t, stderr.String(), "already exists")
	})

	t.Run("requires both paths", func(t *testing.T) {
		stderr.Reset()
		assert.Equal(t, 2, Main([]string{"-from", fixture}, &stdout, &stderr))
		assert.True(t, strings.HasPrefix(stderr.String(), "Usage:"))
	})

	t.Run("removes the target when the import fails", func(t *testing.T) {
		bad := filepath.Join(dir, "not-legacy.db")
		require.NoError(t, os.WriteFile(bad, nil, 0o600))
		target := filepath.Join(dir, "failed.db")
		assert.Equal(t, 1, Main([]string{"-from", bad, "-to", target}, &stdout, &stderr))
		assert.NoFileExists(t, target)
	})
}
