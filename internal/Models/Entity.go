package Models

import (
	"github.com/google/uuid"
)

type Entity struct {
	Base
	Path         string `json:"path"`
	RepositoryID uuid.UUID `gorm:"repository_id constraint:OnDelete:CASCADE;"`
	SnapshotID   string `json:"snapshot_id"`
}
