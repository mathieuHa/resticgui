package Models

import "gorm.io/gorm"

//import "github.com/google/uuid"

type BackupPolicy struct {
	gorm.Model

	Name         string `json:"name" gorm:"unique"`
	NBLastBackup uint   `json:"nb_last_backup"`
	OneByHour    bool   `json:"one_by_hour"`
	OneByDay     bool   `json:"one_by_day"`
	OneByMonth   bool   `json:"one_by_month"`
	KeepAll      bool   `json:"keep_all"`
	RepositoryID uint   // `json:"repository_id"`
}
