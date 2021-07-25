package Models

import (
	"github.com/google/uuid"
)

type Entity struct {
	Base
	Path         string `json:"path"`
	RepositoryId uuid.UUID `gorm:"repository_id" "constraint:OnDelete:CASCADE;"`
	SnapshotId   string `json:"snapshot_id"`
}
