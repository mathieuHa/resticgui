package Models

import "time"

type Snapshot struct {
	Base
	Date         time.Time `gorm:"autoCreateTime"`
	RepositoryID string
	Entities     []Entity
}
