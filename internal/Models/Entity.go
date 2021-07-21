package Models

type Entity struct {
	ID           string `gorm:"primaryKey"`
	Path         string
	RepositoryId string
	SnapshotId   string
}
