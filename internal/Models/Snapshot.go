package Models

import (
	"time"
)

type Snapshot struct {
	Base

	Date         time.Time `json:"date" gorm:"autoCreateTime"`
	RepositoryID uint      //uuid.UUID    `json:"repository_id"`
	Entities     []Entity
}
