package backup

import (
	"errors"
	"os"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/rs/zerolog/log"
)

func validLocalProvider(provider *ent.LocalProvider) (valid bool, err error) {
	log.Debug().Msg("validLocalProvider()")

	if provider == nil {
		valid = false
		err = errors.New("localProvider is nil")
		return valid, err
	}

	log.Info().Msgf("repoPath: %v", provider.RepoPath)
	fileinfo, err := os.Stat(provider.RepoPath)
	// repository will be created by restic in repoPath
	if os.IsNotExist(err) {
		log.Err(err).Msgf("Repository Path %v does not exist", provider.RepoPath)
		return false, err
	}
	if err != nil {
		log.Err(err).Msgf("Repository Path %v unknown error", provider.RepoPath)
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
