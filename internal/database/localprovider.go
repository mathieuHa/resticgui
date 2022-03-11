package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

func UpdateLocalProviderDB(localProvider ent.LocalProvider, ctx context.Context) (b2 *ent.LocalProvider, err error) {
	log.Debug().Msg("UpdateLocalProvider()")

	db := Database.Connect()
	defer db.Close()

	b2, err = db.LocalProvider.
		UpdateOneID(localProvider.ID).
		SetRepoPath(localProvider.RepoPath).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
	}
	return b2, err
}

func CreateLocalProviderDB(localProvider ent.LocalProvider, repoID int, ctx context.Context) (b2 *ent.LocalProvider, err error) {
	log.Debug().Msg("CreateLocalProvider()")

	db := Database.Connect()
	defer db.Close()

	b2, err = db.LocalProvider.
		Create().
		SetRepoPath(localProvider.RepoPath).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return b2, err
}
