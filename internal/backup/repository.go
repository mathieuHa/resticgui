package backup

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/commands"
	"gitlab.com/resticgui/resticgui/internal/database"
	Models "gitlab.com/resticgui/resticgui/internal/models"
	"gitlab.com/resticgui/resticgui/internal/utils"
)

func GetStatsRepository(r *ent.Repository, snapshotID string) (out string, err string) {
	log.Debug().Msg("GetStatsRepository()")

	ctx := context.Background()
	stdout, stderr, errCmd := commands.StatsRepository(r, snapshotID)
	database.LogCommand(r.ID, "StatsSnapshots", stdout, stderr, errCmd, ctx)
	return stdout, stderr
}

func ForgetSnapshotRepository(r *ent.Repository, snapshotID string) (out string, err string) {
	log.Debug().Msg("ForgetSnapshotRepository()")

	ctx := context.Background()
	stdout, stderr, errCmd := commands.ForgetSnapshots(r, snapshotID)
	database.LogCommand(r.ID, "ForgetSnapshots", stdout, stderr, errCmd, ctx)
	return stdout, stderr
}

func CheckRepository(r *ent.Repository) (out string) {
	log.Debug().Msg("CheckRepository()")

	ctx := context.Background()
	stdout, stderr, errCmd := commands.CheckIntegrity(r, false, 0)
	database.LogCommand(r.ID, "CheckIntegrity", stdout, stderr, errCmd, ctx)
	return stdout
}

func RestoreRepository(r *ent.Repository, path string, snapshotID string) (out string, stderr string) {
	log.Debug().Msg("RestoreRepository()")

	ctx := context.Background()

	// TODO Check the validity of the provider and repository in general
	validPath, err := validRestorePath(path)
	if err != nil {
		return "Path is not valid", err.Error()
	}
	if !validPath {
		log.Error().Msg("Not valid")
		return "Path is not valid", ""
	}
	// TODO Check that the snapshot exists

	r, _ = database.GetRepositoryByIDDB(r.ID, ctx)
	// stdout, _, _ := commands.GetSnapshots(r)
	// var snapshots []Models.Snapshot
	// err = json.Unmarshal([]byte(stdout), &snapshots)
	// if err == nil && snapshots != nil {
	// 	r.SnapshotCount = len(snapshots)
	// 	_, err = database.UpdateRepositoryDB(*r, r.ID, ctx)
	// 	if err != nil {
	// 		log.Err(err).Msg("Error update db")
	// 	}
	// }

	// set restore running
	_, _ = database.SetRepositoryRunningDB(r.ID, true, ctx)

	stdout, stderr, errCmd := commands.RestoreRepository(r, path, snapshotID)
	database.LogCommand(r.ID, "RestoreRepository", stdout, stderr, errCmd, ctx)

	// set restore running ended
	_, _ = database.SetRepositoryRunningDB(r.ID, false, ctx)

	return stdout, stderr
}

func BackupRepository(r *ent.Repository, cr *cron.Cron) (out string) {
	log.Debug().Msg("BackupRepository()")

	log.Info().Msgf("JOB B HAS STARTED %v", r.ID)

	ctx := context.Background()

	var cronEntry cron.Entry
	if r.Edges.BackupSchedule != nil && r.Edges.BackupSchedule.CronID != nil {
		cronID := cron.EntryID(*r.Edges.BackupSchedule.CronID)
		cronEntry = cr.Entry(cronID)
	} else {
		log.Error().Msg("Not cron entry")
	}
	// TODO Check the validity of the provider and repository in general

	// Check path len > 0
	if r.Edges.Paths == nil || len(r.Edges.Paths) == 0 {
		log.Info().Msg("Path list is empty, nothing to backup")
		return
	}

	r, _ = database.GetRepositoryByIDDB(r.ID, ctx)
	stdout, stderr, errCmd := commands.GetSnapshots(r)
	var snapshots []Models.Snapshot
	err := json.Unmarshal([]byte(stdout), &snapshots)
	if err == nil && snapshots != nil {
		r.SnapshotCount = len(snapshots)
		_, err = database.UpdateRepositoryDB(*r, r.ID, ctx)
		if err != nil {
			log.Err(err).Msg("Error update db")
		}
	}
	database.LogCommand(r.ID, "GetSnapshots", stdout, stderr, errCmd, ctx)

	// condition for uninitialized repository
	if len(snapshots) == 0 {
		stdout, stderr, errCmd = commands.InitRepository(r)
		database.LogCommand(r.ID, "InitRepository", stdout, stderr, errCmd, ctx)
		// set intiated to true
		// how to check it succeded
	}
	database.SetRepositoryInitiatedDB(r.ID, ctx)

	// set backup running
	_, _ = database.SetRepositoryRunningDB(r.ID, true, ctx)

	_ = database.UpdateCronLastRun(*r.Edges.BackupSchedule, cronEntry.Prev.String())
	_ = database.UpdateCronNextRun(*r.Edges.BackupSchedule, cronEntry.Next.String())

	stdout, stderr, errCmd = commands.BackupRepository(r, false)
	database.LogCommand(r.ID, "BackupRepository", stdout, stderr, errCmd, ctx)

	// set backup running ended
	_, _ = database.SetRepositoryRunningDB(r.ID, false, ctx)

	log.Info().Msgf("JOB B HAS ENDED %v", r.ID)

	return stdout
}

func PruneRepository(r *ent.Repository, cr *cron.Cron) (out string) {
	// It is advisable to run restic check after pruning, to make sure you are alerted, should the internal data structures of the repository be damaged.
	log.Debug().Msg("PruneRepository()")

	log.Info().Msgf("JOB HAS STARTED %v", r.ID)

	ctx := context.Background()

	var cronEntry cron.Entry
	if r.Edges.PruneSchedule != nil && r.Edges.PruneSchedule.CronID != nil {
		cronID := cron.EntryID(*r.Edges.PruneSchedule.CronID)
		cronEntry = cr.Entry(cronID)
	} else {
		log.Error().Msg("Not cron entry")
	}
	// TODO Check the validity of the provider and repository in general

	r, _ = database.GetRepositoryByIDDB(r.ID, ctx)
	stdout, stderr, errCmd := commands.GetSnapshots(r)
	var snapshots []Models.Snapshot
	err := json.Unmarshal([]byte(stdout), &snapshots)
	if err == nil && snapshots != nil {
		r.SnapshotCount = len(snapshots)
		_, err = database.UpdateRepositoryDB(*r, r.ID, ctx)
		if err != nil {
			log.Err(err).Msg("Error update db")
		}
	}
	database.LogCommand(r.ID, "GetSnapshots", stdout, stderr, errCmd, ctx)

	if errCmd != nil {
		stdout, stderr, errCmd = commands.InitRepository(r)
		database.LogCommand(r.ID, "InitRepository", stdout, stderr, errCmd, ctx)
		// set intiated to true
		// how to check it succeded
	}
	database.SetRepositoryInitiatedDB(r.ID, ctx)

	// set prune running
	_, _ = database.SetRepositoryRunningDB(r.ID, true, ctx)

	_ = database.UpdatePruneCronLastRun(*r.Edges.PruneSchedule, cronEntry.Prev.String())
	_ = database.UpdatePruneCronNextRun(*r.Edges.PruneSchedule, cronEntry.Next.String())

	stdout, stderr, errCmd = commands.ForgetSnapshots(r, "")
	database.LogCommand(r.ID, "PruneRepository", stdout, stderr, errCmd, ctx)

	// set prune running ended
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

func ListFilesSnapshotRepository(r *ent.Repository, version string) (out string, stderr string) {
	log.Debug().Msg("ListFilesSnapshotRepository()")
	ctx := context.Background()
	stdout, stderr, errCmd := commands.ListFilesSnapshot(r, version)
	database.LogCommand(r.ID, "ListFiles", stdout, stderr, errCmd, ctx)
	return stdout, stderr
}

func SearchFilesSnapshotRepository(r *ent.Repository, version string, pattern string) (out string, stderr string) {
	log.Debug().Msg("SearchFilesSnapshotRepository()")
	ctx := context.Background()
	stdout, stderr, errCmd := commands.SearchFilesSnapshot(r, version, pattern)
	database.LogCommand(r.ID, "SearchFiles", stdout, stderr, errCmd, ctx)
	return stdout, stderr
}

func DownloadFilesSnapshotRepository(r *ent.Repository, version string, path string, tmpPath string) (out string, stderr string) {
	log.Debug().Msg("DownloadFilesSnapshotRepository()")
	ctx := context.Background()
	stdout, stderr, errCmd := commands.DownloadFilesSnapshot(r, version, path, tmpPath)
	database.LogCommand(r.ID, "DownloadFiles", stdout, stderr, errCmd, ctx)
	return stdout, stderr
}

func ValidRepository(repository ent.Repository) (valid bool, err error) {
	log.Debug().Msg("ValidRepository()")

	valid = (repository.Password != "" && len(repository.Password) > 4)
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
	valid, err = validPaths(repository.Edges.Paths)
	if err != nil {
		log.Err(err).Msg(utils.ErrMsgREmpty_Paths)
		return false, err
	}
	if !valid { // If path is not valid, repository cannot be valid
		return false, nil
	}
	return valid, err
}
