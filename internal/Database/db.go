package Database

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func connect() (_db *gorm.DB) {
	dbname := viper.GetString("db.name")
	_db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		panic("failed to connect database {err} ")
	} else {
		log.Debug().Msgf("Connected to database: %v", dbname)
	}
	return _db
}

func Migrate() {
	db := connect()
	err := db.AutoMigrate(
		&Models.User{},
		&Models.Entity{},
		&Models.BackupPolicy{},
		&Models.Snapshot{},
		&Models.Permission{},
		&Models.Role{},
		&Models.Repository{},
	)
	if err != nil {
		panic("failed to migrate {err} ")
	} else {
		log.Debug().Msg("Auto Migrate")
	}
}
