package Models

type Entity struct {
	Base
	Path         string `json:"path"`
	RepositoryId string `json:"repository_id"`
	SnapshotId   string `json:"snapshot_id"`
}
