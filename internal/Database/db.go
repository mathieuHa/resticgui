package Database

import (
	"github.com/spf13/viper"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"log"
	"os"
)

func Connect() (_db *gorm.DB) {
	dbname := viper.GetString("db.name")
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Error, // Log level Info will output everything
		},
	)
	_db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database {err} ")
	} else {
		//log.Debug().Msgf("Connected to database: %v", dbname)
	}
	return _db
}

func Migrate() {
	db := Connect()
	err := db.AutoMigrate(
		&Models.User{},
		&Models.Path{},
		&Models.BackupPolicy{},
		&Models.BackupSchedule{},
		&Models.Permission{},
		&Models.Role{},
		&Models.Repository{},
	)
	if err != nil {
		panic("failed to migrate {err} ")
	}
}
