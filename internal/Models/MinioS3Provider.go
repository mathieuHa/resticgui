package Models

type MinioS3Provider struct {
	Base

	BucketName   string `json:"bucketName"`
	IP           string `json:"IP"`
	Port         string `json:"port"`
	Scheme       string `json:"scheme"`
	RepositoryID uint   `json:"repositoryID"`
}
