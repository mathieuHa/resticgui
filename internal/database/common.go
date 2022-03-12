package database

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/spf13/viper"
	"gitlab.com/resticgui/resticgui/ent"
)

var db *ent.Client

func Connect() (_db *ent.Client) {
	log.Debug().Msg("Connect()")
	dbname := viper.GetString("db.name")
	client, err := ent.Open("sqlite3", dbname+"?_fk=1")
	if err != nil {
		log.Err(err).Msg("failed opening connection to sqlite")
	}
	return client
}

func Migrate() {
	log.Debug().Msg("Migrate()")
	db = Connect()
	if err := db.Schema.Create(context.Background()); err != nil {
		log.Err(err).Msg("failed creating schema resources")
	}
}

func Close() {
	log.Debug().Msg("Close()")
	db.Close()
}

func CleanDatabaseTest(t assert.TestingT) {
	ctx := context.Background()
	db.Repository.Delete().Exec(ctx)
	db.BackupPolicy.Delete().Exec(ctx)
	db.BackupSchedule.Delete().Exec(ctx)
	db.Path.Delete().Exec(ctx)
	db.AWSS3Provider.Delete().Exec(ctx)
	db.MinioS3Provider.Delete().Exec(ctx)
	db.B2Provider.Delete().Exec(ctx)
	db.LocalProvider.Delete().Exec(ctx)
}
