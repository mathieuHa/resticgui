package Models

import "github.com/google/uuid"

type BackupPolicy struct {
	Base
	Name         string `json:"name" gorm:"unique"`
	NBLastBackup uint   `json:"nb_last_backup"`
	OneByHour    bool   `json:"one_by_hour"`
	OneByDay     bool   `json:"one_by_day"`
	OneByMonth   bool   `json:"one_by_month"`
	KeepAll      bool   `json:"keep_all"`
	RepositoryId uuid.UUID `json:"repository_id"`
}
