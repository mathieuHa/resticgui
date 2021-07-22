package Models

import (
	"time"
)

type Repository struct {
	Base
	Name         string    `json:"name" gorm:"unique"`
	Date         time.Time `json:"date" gorm:"autoCreateTime"`
	Snapshots    []Snapshot
	Entities     []Entity
	BackupPolicy BackupPolicy
}
