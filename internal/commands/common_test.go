package commands

import (
	"context"
	"reflect"
	"testing"

	"github.com/mathieuHa/resticgui/ent"
	Config "github.com/mathieuHa/resticgui/internal/config"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/stretchr/testify/assert"
)

// func TestGetRestoreCommonArgs(t *testing.T) {
// 	t.Run("test latest restore to /", func(t *testing.T) {
// 		path := "/"
// 		version := "latest"
// 		got := getRestoreCommonArgs(path, version)
// 		want := restoreCmd + " latest --target /"
// 		assert.Equal(t, got, want)
// 	})
// }

func Test_getRestoreCommonArgs(t *testing.T) {
	type args struct {
		path    string
		version string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"restore common args", args{"/", "latest"}, []string{"restore", "latest", "--target", "/"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRestoreCommonArgs(tt.args.path, tt.args.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRestoreCommonArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatsRepository(t *testing.T) {
	Config.InitTestConfig()
	database.Migrate()

	t.Run("Stats repository", func(t *testing.T) {
		database.CleanDatabaseTest(t)

		rName := "myRepo"
		repository := ent.Repository{Name: rName}
		r, err := database.CreateRepositoryDB(repository, context.Background())
		assert.Nil(t, err)

		out, _, _ := StatsRepository(r, "latest")
		assert.Equal(t, out, "")
	})
}
