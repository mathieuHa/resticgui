package Models

type B2Provider struct {
	Base

	BucketName   string `json:"bucketName"`
	RepoName     string `json:"repoName"`
	RepositoryID uint   `json:"repositoryID"`
}
