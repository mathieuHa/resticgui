package repository

import (
	"errors"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func validRepository(repository ent.Repository) (valid bool, err error) {
	log.Debug().Msg("validRepository()")

	valid = (repository.Password != "")
	if !valid { // If password is empty, repository cannot be valid
		log.Debug().Msg("Password empty, repository cannot be valid")
		return false, errors.New("Password empty, repository cannot be valid")
	}
	// valid, err = validProvider(repository)
	// if err != nil {
	// 	log.Err(err).Msg("Valid repository has failed on provider check")
	// 	return false, err
	// }
	// if !valid { // If provider is not valid, repository cannot be valid
	// 	return false, nil
	// }
	// valid, err = validSchedule(repository.Edges.BackupSchedule)
	// if err != nil {
	// 	log.Err(err).Msg("Valid repository has failed on schedule check")
	// 	return false, err
	// }
	if !valid { // If schedule is not valid, repository cannot be valid
		return false, nil
	}
	return valid, err
}
