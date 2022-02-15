package Models

type BackupPolicy struct {
	Base

	Name         string `json:"name" gorm:"unique"`
	NBLastBackup uint   `json:"nbLastBackup"`
	OneByHour    bool   `json:"oneByHour"`
	OneByDay     bool   `json:"oneByDay"`
	OneByMonth   bool   `json:"oneByMonth"`
	KeepAll      bool   `json:"keepAll"`
	RepositoryID uint   `json:"repositoryID"`
}
