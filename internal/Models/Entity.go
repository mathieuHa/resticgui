package Models

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Entity struct {
	Base
	Path         string    `json:"path"`
	RepositoryID uuid.UUID `gorm:"repository_id constraint:OnDelete:CASCADE;"`
	SnapshotID   uuid.UUID  `json:"snapshot_id"`
	Exists       bool		`json:"exists"`
}

func (e Entity) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Path)
}

