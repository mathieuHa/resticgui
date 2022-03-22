package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Config"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func TestGetProviderData(t *testing.T) {
	Config.InitTestConfig()
	database.Migrate()

	t.Run("Stats repository", func(t *testing.T) {
		database.CleanDatabaseTest(t)

		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)

		out, _ := getProviderData(*r)
		assert.Equal(t, out, "")

	})
}
