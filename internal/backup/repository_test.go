package backup

import (
	"bytes"
	"context"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/config"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func TestGetStatsRepository(t *testing.T) {
	Config.InitTestConfig()
	database.Migrate()

	t.Run("get stats repository", func(t *testing.T) {
		database.CleanDatabaseTest(t)

		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)

		out, _ := GetStatsRepository(r, "latest")
		assert.Equal(t, out, "")

	})

}

// regression: BackupRepository used to log the repository password at debug level
func TestBackupRepositoryDoesNotLogPassword(t *testing.T) {
	Config.InitTestConfig()
	database.Migrate()
	database.CleanDatabaseTest(t)
	ctx := context.Background()

	var buf bytes.Buffer
	previousLogger, previousLevel := log.Logger, zerolog.GlobalLevel()
	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	t.Cleanup(func() {
		log.Logger = previousLogger
		zerolog.SetGlobalLevel(previousLevel)
	})

	const password = "do-not-log-this-password"
	r, err := database.CreateRepositoryDB(ent.Repository{Name: "secretRepo", Password: password, Active: true}, ctx)
	assert.Nil(t, err)
	_, err = database.CreateLocalProviderDB(ent.LocalProvider{RepoPath: t.TempDir()}, r.ID, ctx)
	assert.Nil(t, err)
	_, err = database.CreateBackupScheduleDB(ent.BackupSchedule{Choice: "interval", Interval: "1h"}, r.ID, ctx)
	assert.Nil(t, err)
	src := t.TempDir()
	_, err = database.CreatePathsDB([]*ent.Path{{Name: "src", AbsPath: src, RelPath: src, IsDir: true, Included: true}}, r.ID, ctx)
	assert.Nil(t, err)
	_, err = database.SetRepositoryValidDB(true, r.ID, ctx)
	assert.Nil(t, err)

	BackupRepository(r.ID, cron.New())

	assert.Contains(t, buf.String(), "BackupRepository()", "debug logs should have been captured")
	assert.NotContains(t, buf.String(), password)
}
