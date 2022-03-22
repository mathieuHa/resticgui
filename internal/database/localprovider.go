package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func UpdateLocalProviderDB(localProvider ent.LocalProvider, id int, ctx context.Context) (lp *ent.LocalProvider, err error) {
	log.Debug().Msg("UpdateLocalProviderDB()")

	lp, err = db.LocalProvider.
		UpdateOneID(id).
		SetRepoPath(localProvider.RepoPath).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errUpdate)
	}
	return lp, err
}

func CreateLocalProviderDB(localProvider ent.LocalProvider, repoID int, ctx context.Context) (lp *ent.LocalProvider, err error) {
	log.Debug().Msg("CreateLocalProviderDB()")

	lp, err = db.LocalProvider.
		Create().
		SetRepoPath(localProvider.RepoPath).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return lp, err
}
