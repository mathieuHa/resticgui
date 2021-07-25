package Models

import (
	"github.com/google/uuid"
	"time"
)

type Snapshot struct {
	Base
	Date         time.Time `json:"date" gorm:"autoCreateTime"`
	RepositoryID uuid.UUID    `json:"repository_id"`
	Entities     []Entity
}
