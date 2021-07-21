package Models

type BackupPolicy struct {
	ID           string
	Name         string `gorm:"unique"`
	NBLastBackup uint
	OneByHour    bool
	OneByDay     bool
	OneByMonth   bool
	KeepAll      bool
	RepositoryId string
}
