package Models

type BackupPolicy struct {
	Base
	Name         string `gorm:"unique"`
	NBLastBackup uint
	OneByHour    bool
	OneByDay     bool
	OneByMonth   bool
	KeepAll      bool
	RepositoryId string
}
