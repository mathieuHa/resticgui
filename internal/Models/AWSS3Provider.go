package Models

type AWSS3Provider struct {
	Base

	BucketName   string `json:"bucketName"`
	RepositoryID uint   `json:"repositoryID"`
}
