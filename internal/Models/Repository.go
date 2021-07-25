package Models

import (
	"time"
)

type Repository struct {
	Base
	Name         string         `json:"name" gorm:"unique"`
	Date         time.Time      `json:"date" gorm:"autoCreateTime"`
	Snapshots    []Snapshot     `gorm: "constraint:OnDelete:CASCADE;"`
	Entities     []Entity       `gorm: "constraint:OnDelete:CASCADE;"`
	BackupPolicy BackupPolicy
	BackupSchedule BackupSchedule
}
