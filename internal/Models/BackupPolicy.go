package Models

type BackupPolicy struct {
	Base

	Name         string `json:"name" gorm:"unique"`
	NBLastBackup uint   `json:"nb_last_backup"`
	OneByHour    bool   `json:"one_by_hour"`
	OneByDay     bool   `json:"one_by_day"`
	OneByMonth   bool   `json:"one_by_month"`
	KeepAll      bool   `json:"keep_all"`
	RepositoryID uint   // `json:"repository_id"`
}
