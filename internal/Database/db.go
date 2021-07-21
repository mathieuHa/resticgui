package Database

import (
	"fmt"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func connect() (_db *gorm.DB) {
	_db, err := gorm.Open(sqlite.Open("resticgui.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database {err} ")
	} else {
		fmt.Println("Connected to database")
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
		fmt.Println("Auto Migrate")
	}
}
