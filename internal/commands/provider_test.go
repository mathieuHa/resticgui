package commands

import (
	"context"
	"testing"

	"github.com/mathieuHa/resticgui/ent"
	Config "github.com/mathieuHa/resticgui/internal/config"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestGetProviderData(t *testing.T) {
	Config.InitTestConfig()
	database.Migrate()

	t.Run("Get Provider Data", func(t *testing.T) {
		database.CleanDatabaseTest(t)

		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		if err == nil {
			out, _, _ := getProviderData(r)
			assert.Equal(t, []string([]string(nil)), out)
		}
	})
}

func TestGetLocalRepositoryFormat(t *testing.T) {
	Config.InitTestConfig()
	database.Migrate()

	t.Run("Get local repository format", func(t *testing.T) {
		database.CleanDatabaseTest(t)

		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)
		id := r.ID
		myPath := "myPath"
		p := ent.LocalProvider{RepoPath: myPath}
		pa, err := database.CreateLocalProviderDB(p, id, context.Background())
		assert.Nil(t, err)
		if err == nil {
			out := getLocalRepositoryFormat(*pa)
			assert.Equal(t, out, myPath)
		}
	})
}
