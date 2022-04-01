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
		log.Debug().Msg("password empty, repository cannot be valid")
		return false, errors.New("password empty, repository cannot be valid")
	}
	valid, err = validProvider(&repository)
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
