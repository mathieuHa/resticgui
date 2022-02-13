package Models

import (
	"github.com/rs/zerolog/log"
)

type Entity struct {
	Base

	Path         string `json:"path" binding:"required"`
	RepositoryID uint   `gorm:"repository_id constraint:OnDelete:CASCADE;"`
	SnapshotID   uint   //uuid.UUID  `json:"snapshot_id"`
	Exists       bool   `json:"exists"`
}

func (e Entity) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Path)
}
