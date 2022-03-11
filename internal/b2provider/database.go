package b2provider

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

func UpdateB2ProviderDB(b2Provider ent.B2Provider, ctx context.Context) (b2 *ent.B2Provider, err error) {
	log.Debug().Msg("UpdateB2Provider()")

	db := Database.Connect()
	defer db.Close()

	b2, err = db.B2Provider.
		UpdateOneID(b2Provider.ID).
		SetAccountID(b2Provider.AccountID).
		SetAccountKey(b2Provider.AccountKey).
		SetBucketName(b2Provider.BucketName).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
	}
	return b2, err
}

func CreateB2ProviderDB(b2Provider ent.B2Provider, repoID int, ctx context.Context) (b2 *ent.B2Provider, err error) {
	log.Debug().Msg("CreateB2Provider()")

	db := Database.Connect()
	defer db.Close()

	b2, err = db.B2Provider.
		Create().
		SetAccountID(b2Provider.AccountID).
		SetAccountKey(b2Provider.AccountKey).
		SetBucketName(b2Provider.BucketName).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
	}

	return b2, err
}
