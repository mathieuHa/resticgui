package Models

import (
	"github.com/rs/zerolog/log"
)

type Repository struct {
	Base

	Name           string         `json:"name" gorm:"unique" binding:"required,min=3"`
	Paths   []Path         `json:"paths" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BackupPolicy   BackupPolicy   `json:"backupPolicy" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BackupSchedule BackupSchedule `json:"backupSchedule" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Active         bool           `json:"active"`
}

func (e Repository) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Name)
}
