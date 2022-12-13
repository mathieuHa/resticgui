package backup

import (
	"context"
	"testing"

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
