package API

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func updateAWSS3Provider(db *ent.Client, awsS3Provider ent.AWSS3Provider, ctx context.Context) {
	_, err := db.AWSS3Provider.
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
}

func createAWSS3Provider(db *ent.Client, awsS3Provider ent.AWSS3Provider, repoID int, ctx context.Context) {
	_, err := db.AWSS3Provider.
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
}
