package API

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func updateLocalProvider(db *ent.Client, localProvider ent.LocalProvider, ctx context.Context) {

}

func createLocalProvider(db *ent.Client, localProvider ent.LocalProvider, repoID int, ctx context.Context) {
	_, err := db.LocalProvider.
		Create().
		SetRepoPath(localProvider.RepoPath).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
}
