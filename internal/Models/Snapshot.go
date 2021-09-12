package Models

import (
	//"github.com/google/uuid"
	"time"

	"gorm.io/gorm"
)

type Snapshot struct {
	gorm.Model

	Date         time.Time `json:"date" gorm:"autoCreateTime"`
	RepositoryID uint      //uuid.UUID    `json:"repository_id"`
	Entities     []Entity
}
