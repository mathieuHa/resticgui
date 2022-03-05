package API

import (
	"context"

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
