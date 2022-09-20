package backup

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/commands"
	"gitlab.com/resticgui/resticgui/internal/database"
	"gitlab.com/resticgui/resticgui/internal/utils"
)

func GetStatsRepository(r *ent.Repository) (out string) {
	log.Debug().Msg("GetStatsRepository()")

	ctx := context.Background()
	stdout, stderr, errCmd := commands.StatsRepository(r, false)
	database.LogCommand(r.ID, "StatsSnapshots", stdout, stderr, errCmd, ctx)
	return stdout
}

func CheckRepository(r *ent.Repository) (out string) {
	log.Debug().Msg("CheckRepository()")

	ctx := context.Background()
	stdout, stderr, errCmd := commands.CheckIntegrity(r, false, 0)
	database.LogCommand(r.ID, "CheckIntegrity", stdout, stderr, errCmd, ctx)
	return stdout
}

func BackupRepository(r *ent.Repository) (out string) {
	log.Debug().Msg("BackupRepository()")

	log.Info().Msgf("JOB HAS STARTED %v", r.ID)

	ctx := context.Background()
	// TODO Check the validity of the provider and repository in general

	r, _ = database.GetRepositoryByIDDB(r.ID, ctx)
	stdout, stderr, errCmd := commands.GetSnapshots(r)
	database.LogCommand(r.ID, "GetSnapshots", stdout, stderr, errCmd, ctx)

	if errCmd != nil {
		stdout, stderr, errCmd = commands.InitRepository(r)
		database.LogCommand(r.ID, "InitRepository", stdout, stderr, errCmd, ctx)
		// set intiated to true
		// how to check it succeded
		database.SetRepositoryInitiatedDB(r.ID, ctx)
	}

	// set backup running
	_, _ = database.SetRepositoryRunningDB(r.ID, true, ctx)

	stdout, stderr, errCmd = commands.BackupRepository(r, false)
	database.LogCommand(r.ID, "BackupRepository", stdout, stderr, errCmd, ctx)

	// set backup running ended
	_, _ = database.SetRepositoryRunningDB(r.ID, false, ctx)

	log.Info().Msgf("JOB HAS ENDED %v", r.ID)

	return stdout
}

func GetSnapshotsRepository(r *ent.Repository) (out string) {
	log.Debug().Msg("GetSnapshotsRepository()")

	ctx := context.Background()
	stdout, stderr, errCmd := commands.GetSnapshots(r)
	database.LogCommand(r.ID, "GetSnapshots", stdout, stderr, errCmd, ctx)
	return stdout
}

func ValidRepository(repository ent.Repository) (valid bool, err error) {
	log.Debug().Msg("ValidRepository()")

	valid = (repository.Password != "")
	if !valid { // If password is empty, repository cannot be valid
		log.Debug().Msg(utils.ErrMsgRInvalid_Password)

		return false, errors.New(utils.ErrMsgRInvalid_Password)
	}
	valid, err = validProvider(&repository)
	if err != nil {
		log.Err(err).Msg(utils.ErrMsgRInvalid_Provider)
		return false, err
	}
	if !valid { // If provider is not valid, repository cannot be valid
		return false, nil
	}
	valid, err = validSchedule(repository.Edges.BackupSchedule)
	if err != nil {
		log.Err(err).Msg(utils.ErrMsgRInvalid_Schedule)
		return false, err
	}
	if !valid { // If schedule is not valid, repository cannot be valid
		return false, nil
	}
	return valid, err
}
