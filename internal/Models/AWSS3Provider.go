package Models

type AWSS3Provider struct {
	Base

	BucketName      string `json:"bucketName"`
	RepositoryID    uint   `json:"repositoryID"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	DefaultRegion   string `json:"defaultRegion"`
}
