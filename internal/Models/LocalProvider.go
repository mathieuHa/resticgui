package Models

type LocalProvider struct {
	Base

	RepoName     string `json:"repoName"`
	RepositoryID uint   `json:"repositoryID"`
}
