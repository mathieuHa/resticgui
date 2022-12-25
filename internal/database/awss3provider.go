package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func UpdateAWSS3ProviderDB(s3Provider ent.AWSS3Provider, id int, ctx context.Context) (s3 *ent.AWSS3Provider, err error) {
	log.Debug().Msg("UpdateAWSS3Provider()")

	s3, err = db.AWSS3Provider.
		UpdateOneID(id).
		SetAccessKeyID(s3Provider.AccessKeyID).
		SetSecretAccessKey(s3Provider.SecretAccessKey).
		SetBucketName(s3Provider.BucketName).
		SetDefaultRegion(s3Provider.DefaultRegion).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
	}
	return s3, err
}

func CreateAWSS3ProviderDB(s3Provider ent.AWSS3Provider, repoID int, ctx context.Context) (s3 *ent.AWSS3Provider, err error) {
	log.Debug().Msg("CreateAWSS3Provider()")

	s3, err = db.AWSS3Provider.
		Create().
		SetAccessKeyID(s3Provider.AccessKeyID).
		SetSecretAccessKey(s3Provider.SecretAccessKey).
		SetBucketName(s3Provider.BucketName).
		SetDefaultRegion(s3Provider.DefaultRegion).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return s3, err
}
