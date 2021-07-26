package Models

import (
	"time"
)

type Repository struct {
	Base
	Name         string         	`json:"name" gorm:"unique"`
	Date         time.Time      	`json:"date" gorm:"autoCreateTime"`
	Snapshots    []Snapshot     	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Entities     []Entity       	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BackupPolicy BackupPolicy		`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BackupSchedule BackupSchedule	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

