package Models

import (
	//"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Entity struct {
	gorm.Model

	Path         string `json:"path"`
	RepositoryID uint   //uuid.UUID `gorm:"repository_id constraint:OnDelete:CASCADE;"`
	SnapshotID   uint   //uuid.UUID  `json:"snapshot_id"`
	Exists       bool   `json:"exists"`
}

func (e Entity) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Path)
}
