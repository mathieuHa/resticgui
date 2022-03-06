package API

import (
	"context"
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func updateLocalProvider(db *ent.Client, localProvider *ent.LocalProvider, ctx context.Context) (err error) {
	log.Debug().Msg("updateLocalProvider()")

	_, err = db.LocalProvider.
		UpdateOneID(localProvider.ID).
		SetRepoPath(localProvider.RepoPath).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func createLocalProvider(db *ent.Client, localProvider *ent.LocalProvider, repoID int, ctx context.Context) (err error) {
	log.Debug().Msg("createLocalProvider()")

	_, err = db.LocalProvider.
		Create().
		SetRepoPath(localProvider.RepoPath).
		SetRepositoryID(repoID).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}
	return err
}

func validLocalProvider(provider *ent.LocalProvider) (valid bool, err error) {
	log.Debug().Msg("validLocalProvider()")

	if provider == nil {
		valid = false
		err = errors.New("localProvider is nil")
		return valid, err
	}

	log.Info().Msgf("repoPath: %v", provider.RepoPath)
	fileinfo, err := os.Stat(provider.RepoPath)
	if os.IsNotExist(err) {
		log.Err(err).Msgf("Repository Path %v does not exist", provider.RepoPath)
		return false, err
	}
	if !fileinfo.IsDir() {
		log.Err(err).Msgf("Repository Path %v is not a directory", provider.RepoPath)
		return false, err
	}
	if fileinfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		log.Err(err).Msgf("Repository Path %v is not writable by user", provider.RepoPath)
		return false, err
	}

	return true, nil
}
