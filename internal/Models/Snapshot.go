package Models

import "time"

type Snapshot struct {
	Base
	Date         time.Time `json:"date" gorm:"autoCreateTime"`
	RepositoryID string    `json:"repository_id"`
	Entities     []Entity
}
