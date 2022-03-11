package awss3provider

import (
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func validAWSS3Provider(provider *ent.AWSS3Provider) (valid bool, err error) {
	log.Debug().Msg("validAWSS3Provider()")

	if provider == nil {
		valid = false
		err = errors.New("AWSS3Provider is nil")
		return valid, err
	}
	// accessKeyID := provider.AccessKeyID
	// secretAccessKey := provider.SecretAccessKey
	// region := provider.DefaultRegion
	// bucketName := provider.BucketName
	// useSSL := true

	// client := s3.New(s3.Options{
	// 	Region:      region,
	// 	Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentials(accessKeyID, secretAccessKey)),
	// })

	return true, nil
}

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}
