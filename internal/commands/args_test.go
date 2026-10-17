package commands

import (
	"testing"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/stretchr/testify/assert"
)

func Test_getPruneCommonArgs(t *testing.T) {
	tests := []struct {
		name   string
		policy ent.BackupPolicy
		want   []string
	}{
		{"empty policy", ent.BackupPolicy{}, nil},
		{"keep last only", ent.BackupPolicy{NbKeepLast: 10}, []string{"--keep-last", "10"}},
		// regression: --keep-hourly used to receive "true" instead of the count
		{"keep hourly", ent.BackupPolicy{NbKeepHours: 24}, []string{"--keep-hourly", "24"}},
		{"full policy", ent.BackupPolicy{NbKeepLast: 3, NbKeepHours: 6, NbKeepDays: 7, NbKeepWeeks: 4, NbKeepMonths: 12},
			[]string{"--keep-last", "3", "--keep-hourly", "6", "--keep-daily", "7", "--keep-weekly", "4", "--keep-monthly", "12"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := tt.policy
			r := &ent.Repository{Edges: ent.RepositoryEdges{BackupPolicy: &policy}}
			assert.Equal(t, tt.want, getPruneCommonArgs(r))
		})
	}
}

func Test_getDryRunParameter(t *testing.T) {
	assert.Equal(t, []string{"--dry-run"}, getDryRunParameter(true))
	assert.Empty(t, getDryRunParameter(false))
}

func localRepository() *ent.Repository {
	return &ent.Repository{
		Password: "a-strong-password",
		Provider: "localProvider",
		Edges: ent.RepositoryEdges{
			LocalProvider: &ent.LocalProvider{RepoPath: "/srv/restic"},
			Paths: []*ent.Path{
				{RelPath: "/data", Included: true},
				{RelPath: "/data/cache", Included: false},
			},
		},
	}
}

func Test_getBackupArgs(t *testing.T) {
	t.Run("real backup", func(t *testing.T) {
		args, env := getBackupArgs(localRepository(), false)
		assert.Equal(t, []string{"-r", "/srv/restic", "backup", "/data", "--exclude", "/data/cache", "--json"}, args)
		assert.Contains(t, env, "RESTIC_PASSWORD=a-strong-password")
	})

	t.Run("dry run", func(t *testing.T) {
		args, _ := getBackupArgs(localRepository(), true)
		assert.Contains(t, args, "--dry-run")
	})

	// regression: a non dry-run backup used to get an empty "" argument appended
	t.Run("no empty argument", func(t *testing.T) {
		for _, dryRun := range []bool{true, false} {
			args, _ := getBackupArgs(localRepository(), dryRun)
			assert.NotContains(t, args, "")
		}
	})

	t.Run("password never in argv", func(t *testing.T) {
		args, _ := getBackupArgs(localRepository(), false)
		for _, a := range args {
			assert.NotContains(t, a, "a-strong-password")
		}
	})
}
