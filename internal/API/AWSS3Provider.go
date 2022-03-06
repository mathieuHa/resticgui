package API

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func updateAWSS3Provider(db *ent.Client, awsS3Provider *ent.AWSS3Provider, ctx context.Context) (err error) {
	_, err = db.AWSS3Provider.
		UpdateOneID(awsS3Provider.ID).
		SetAccessKeyID(awsS3Provider.AccessKeyID).
		SetSecretAccessKey(awsS3Provider.SecretAccessKey).
		SetBucketName(awsS3Provider.BucketName).
		SetDefaultRegion(awsS3Provider.DefaultRegion).
		Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func createAWSS3Provider(db *ent.Client, awsS3Provider *ent.AWSS3Provider, repoID int, ctx context.Context) (err error) {
	_, err = db.AWSS3Provider.
		Create().
		SetAccessKeyID(awsS3Provider.AccessKeyID).
		SetSecretAccessKey(awsS3Provider.SecretAccessKey).
		SetBucketName(awsS3Provider.BucketName).
		SetDefaultRegion(awsS3Provider.DefaultRegion).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

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
