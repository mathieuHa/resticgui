package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func UpdateMinioS3ProviderDB(minioS3Provider ent.MinioS3Provider, ctx context.Context) (mS3p *ent.MinioS3Provider, err error) {
	log.Debug().Msg("UpdateMinioS3Provider()")

	db := Connect()
	defer db.Close()

	mS3p, err = db.MinioS3Provider.
		UpdateOneID(minioS3Provider.ID).
		SetAccessKeyID(minioS3Provider.AccessKeyID).
		SetSecretAccessKey(minioS3Provider.SecretAccessKey).
		SetBucketName(minioS3Provider.BucketName).
		SetIP(minioS3Provider.IP).
		SetPort(minioS3Provider.Port).
		SetScheme(minioS3Provider.Scheme).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
	}
	return mS3p, err
}

func CreateMinioS3ProviderDB(minioS3Provider ent.MinioS3Provider, repoID int, ctx context.Context) (mS3p *ent.MinioS3Provider, err error) {
	log.Debug().Msg("CreateMinioS3Provider()")

	db := Connect()
	defer db.Close()

	mS3p, err = db.MinioS3Provider.
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
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return mS3p, err
}
