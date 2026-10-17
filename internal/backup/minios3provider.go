package backup

import (
	"errors"
	"net/http"
	"time"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/rs/zerolog/log"
)

func validMinioS3Provider(provider *ent.MinioS3Provider) (valid bool, err error) {
	log.Debug().Msg("validMinioS3Provider()")

	if provider == nil {
		valid = false
		err = errors.New("MinioS3Provider is nil")
		return valid, err
	}
	// endpoint := fmt.Sprintf("%v://%v:%v", provider.Scheme, provider.IP, provider.Port)
	// accessKeyID := provider.AccessKeyID
	// secretAccessKey := provider.SecretAccessKey
	// bucketName := provider.BucketName
	// useSSL := true

	// // Initialize minio client object.
	// minioClient, err := minio.New(endpoint, &minio.Options{
	// 	Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	// 	Secure: useSSL,
	// })
	// if err != nil {
	// 	log.Err(err)
	// 	return false, err
	// }
	// bucketExist, err := minioClient.BucketExists(context.Background(), bucketName)
	// if err != nil {
	// 	log.Err(err)
	// 	return false, err
	// }
	// if !bucketExist {
	// 	return false, nil
	// }
	return true, nil
}

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}
