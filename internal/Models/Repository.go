package Models

import (
	"time"
)

type Repository struct {
	ID           string
	Name         string `gorm:"unique"`
	Date         time.Time
	Snapshots    []Snapshot
	Entities     []Entity
	BackupPolicy BackupPolicy
}
