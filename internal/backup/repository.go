package backup

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/commands"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func GetStatsRepository(r *ent.Repository) (out string) {
	ctx := context.Background()
	stdout, stderr, errCmd := commands.StatsRepository(*r, false)
	database.LogCommand(r.ID, "StatsSnapshots", stdout, stderr, errCmd, ctx)
	return stdout
}

func ValidRepository(repository ent.Repository) (valid bool, err error) {
	log.Debug().Msg("validRepository()")

	valid = (repository.Password != "")
	if !valid { // If password is empty, repository cannot be valid
		log.Debug().Msg("Password empty, repository cannot be valid")
		return false, errors.New("Password empty, repository cannot be valid")
	}
	valid, err = validProvider(repository)
	if err != nil {
		log.Err(err).Msg("Valid repository has failed on provider check")
		return false, err
	}
	if !valid { // If provider is not valid, repository cannot be valid
		return false, nil
	}
	valid, err = validSchedule(repository.Edges.BackupSchedule)
	if err != nil {
		log.Err(err).Msg("Valid repository has failed on schedule check")
		return false, err
	}
	if !valid { // If schedule is not valid, repository cannot be valid
		return false, nil
	}
	return valid, err
}
