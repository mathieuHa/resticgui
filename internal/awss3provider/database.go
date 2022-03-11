package awss3provider

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

func UpdateAWSS3ProviderDB(s3Provider ent.AWSS3Provider, ctx context.Context) (s3 *ent.AWSS3Provider, err error) {
	log.Debug().Msg("UpdateAWSS3Provider()")

	db := Database.Connect()
	defer db.Close()

	s3, err = db.AWSS3Provider.
		UpdateOneID(s3Provider.ID).
		SetAccessKeyID(s3Provider.AccessKeyID).
		SetSecretAccessKey(s3Provider.SecretAccessKey).
		SetBucketName(s3Provider.BucketName).
		SetDefaultRegion(s3Provider.DefaultRegion).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
	}
	return s3, err
}

func CreateAWSS3ProviderDB(s3Provider ent.AWSS3Provider, repoID int, ctx context.Context) (s3 *ent.AWSS3Provider, err error) {
	log.Debug().Msg("CreateAWSS3Provider()")

	db := Database.Connect()
	defer db.Close()

	s3, err = db.AWSS3Provider.
		Create().
		SetAccessKeyID(s3Provider.AccessKeyID).
		SetSecretAccessKey(s3Provider.SecretAccessKey).
		SetBucketName(s3Provider.BucketName).
		SetDefaultRegion(s3Provider.DefaultRegion).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
	}

	return s3, err
}
