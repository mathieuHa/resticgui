package API

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func updateMinioS3Provider(db *ent.Client, minioS3Provider *ent.MinioS3Provider, ctx context.Context) (err error) {
	log.Debug().Msg("updateMinioS3Provider()")
	_, err = db.MinioS3Provider.
		UpdateOneID(minioS3Provider.ID).
		SetAccessKeyID(minioS3Provider.AccessKeyID).
		SetSecretAccessKey(minioS3Provider.SecretAccessKey).
		SetBucketName(minioS3Provider.BucketName).
		SetIP(minioS3Provider.IP).
		SetPort(minioS3Provider.Port).
		SetScheme(minioS3Provider.Scheme).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func createMinioS3Provider(db *ent.Client, minioS3Provider *ent.MinioS3Provider, repoID int, ctx context.Context) (err error) {
	log.Debug().Msg("createMinioS3Provider()")
	_, err = db.MinioS3Provider.
		Create().
		SetAccessKeyID(minioS3Provider.AccessKeyID).
		SetSecretAccessKey(minioS3Provider.SecretAccessKey).
		SetBucketName(minioS3Provider.BucketName).
		SetIP(minioS3Provider.IP).
		SetPort(minioS3Provider.Port).
		SetScheme(minioS3Provider.Scheme).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func validMinioS3Provider(provider *ent.MinioS3Provider) (valid bool, err error) {
	log.Debug().Msg("validMinioS3Provider()")

	if provider == nil {
		valid = false
		err = errors.New("MinioS3Provider is nil")
		return valid, err
	}
	endpoint := fmt.Sprintf("%v://%v:%v", provider.Scheme, provider.IP, provider.Port)
	accessKeyID := provider.AccessKeyID
	secretAccessKey := provider.SecretAccessKey
	bucketName := provider.BucketName
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Err(err)
		return false, err
	}
	bucketExist, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Err(err)
		return false, err
	}
	if !bucketExist {
		return false, nil
	}
	return true, nil
}
