package Models

import "time"

type Snapshot struct {
	ID           string    `gorm:"primaryKey"`
	Date         time.Time `gorm:"autoCreateTime"`
	RepositoryID string
	Entities     []Entity
}
