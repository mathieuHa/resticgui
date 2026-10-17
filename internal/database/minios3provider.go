package database

import (
	"context"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/rs/zerolog/log"
)

func UpdateMinioS3ProviderDB(minioS3Provider ent.MinioS3Provider, id int, ctx context.Context) (mS3p *ent.MinioS3Provider, err error) {
	log.Debug().Msg("UpdateMinioS3Provider()")

	mS3p, err = db.MinioS3Provider.
		UpdateOneID(id).
		SetAccessKeyID(minioS3Provider.AccessKeyID).
		SetSecretAccessKey(minioS3Provider.SecretAccessKey).
		SetBucketName(minioS3Provider.BucketName).
		SetHostname(minioS3Provider.Hostname).
		SetPort(minioS3Provider.Port).
		SetScheme(minioS3Provider.Scheme).
		SetCertificateAuthority(minioS3Provider.CertificateAuthority).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
	}
	return mS3p, err
}

func CreateMinioS3ProviderDB(minioS3Provider ent.MinioS3Provider, repoID int, ctx context.Context) (mS3p *ent.MinioS3Provider, err error) {
	log.Debug().Msg("CreateMinioS3Provider()")

	mS3p, err = db.MinioS3Provider.
		Create().
		SetAccessKeyID(minioS3Provider.AccessKeyID).
		SetSecretAccessKey(minioS3Provider.SecretAccessKey).
		SetBucketName(minioS3Provider.BucketName).
		SetHostname(minioS3Provider.Hostname).
		SetPort(minioS3Provider.Port).
		SetScheme(minioS3Provider.Scheme).
		SetCertificateAuthority(minioS3Provider.CertificateAuthority).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return mS3p, err
}
