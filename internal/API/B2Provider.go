package API

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func updateB2Provider(db *ent.Client, b2Provider ent.B2Provider, ctx context.Context) {
	_, err := db.B2Provider.
		UpdateOneID(b2Provider.ID).
		SetAccountID(b2Provider.AccountID).
		SetAccountKey(b2Provider.AccountKey).
		SetBucketName(b2Provider.BucketName).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
}

func createB2Provider(db *ent.Client, b2Provider ent.B2Provider, repoID int, ctx context.Context) {
	_, err := db.B2Provider.
		Create().
		SetAccountID(b2Provider.AccountID).
		SetAccountKey(b2Provider.AccountKey).
		SetBucketName(b2Provider.BucketName).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
}
