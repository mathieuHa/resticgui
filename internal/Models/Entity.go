package Models

type Entity struct {
	Base
	Path         string `json:"path"`
	RepositoryId string `gorm:"repository_id" "constraint:OnDelete:CASCADE;"`
	SnapshotId   string `json:"snapshot_id"`
}
