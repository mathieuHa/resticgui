package Models

import "gitlab.com/resticgui/resticgui/ent"

type Provider struct {
	Provider        string              `json:"provider"`
	LocalProvider   ent.LocalProvider   `json:"localProvider"`
	B2Provider      ent.B2Provider      `json:"b2Provider"`
	MinioS3Provider ent.MinioS3Provider `json:"minioS3Provider"`
	AWSS3Provider   ent.AWSS3Provider   `json:"awsS3Provider"`
}
