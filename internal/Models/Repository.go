package Models

import (
	"time"
)

type Repository struct {
	ID           string    `gorm:"primaryKey"`
	Name         string    `gorm:"unique"`
	Date         time.Time `gorm:"autoCreateTime"`
	Snapshots    []Snapshot
	Entities     []Entity
	BackupPolicy BackupPolicy
}
