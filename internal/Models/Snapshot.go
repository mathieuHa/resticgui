package Models

import "time"

type Snapshot struct {
	ID           string
	Date         time.Time
	RepositoryID string
	Entities     []Entity
}
