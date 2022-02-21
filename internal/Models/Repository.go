package Models

import (
	"github.com/rs/zerolog/log"
)

type Repository struct {
	Base

	Name            string          `json:"name" gorm:"unique" binding:"required,min=3"`
	Local           string          `json:"local"`
	MaxSize         string          `json:"maxSize"`
	Provider        string          `json:"provider"`
	Password        string          `json:"password"`
	Paths           []Path          `json:"paths" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LocalProvider   LocalProvider   `json:"localProvider" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	B2Provider      B2Provider      `json:"B2Provider" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AWSS3Provider   AWSS3Provider   `json:"AWSS3Provider" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MinioS3Provider MinioS3Provider `json:"MinioS3Provider" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BackupPolicy    BackupPolicy    `json:"backupPolicy" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BackupSchedule  BackupSchedule  `json:"backupSchedule" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Active          bool            `json:"active"`
}

func (e Repository) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Name)
}
