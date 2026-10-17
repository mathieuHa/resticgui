package database

import (
	"context"

	"github.com/mathieuHa/resticgui/ent"
	"github.com/mathieuHa/resticgui/ent/repository"
	"github.com/rs/zerolog/log"
)

const errCreate = "creation failed"
const errUpdate = "update failed"

func CreateBackupScheduleDB(backupSchedule ent.BackupSchedule, repoID int, ctx context.Context) (b *ent.BackupSchedule, err error) {
	log.Debug().Msg("CreateBackupScheduleDB()")

	b, err = db.BackupSchedule.
		Create().
		SetChoice(backupSchedule.Choice).
		SetCronSyntax(backupSchedule.CronSyntax).
		SetDay(backupSchedule.Day).
		SetHour(backupSchedule.Hour).
		SetEveryDay(backupSchedule.EveryDay).
		SetEveryWeek(backupSchedule.EveryWeek).
		SetEveryMonth(backupSchedule.EveryMonth).
		SetEveryMonday(backupSchedule.EveryMonday).
		SetEveryTuesday(backupSchedule.EveryTuesday).
		SetEveryWednesday(backupSchedule.EveryWednesday).
		SetEveryThursday(backupSchedule.EveryThursday).
		SetEveryFriday(backupSchedule.EveryFriday).
		SetEverySaturday(backupSchedule.EverySaturday).
		SetEverySunday(backupSchedule.EverySunday).
		SetInterval(backupSchedule.Interval).
		SetRepositoryID(repoID).
		Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return b, err
}

func UpdateBackupScheduleDB(backupSchedule ent.BackupSchedule, id int, ctx context.Context) (b *ent.BackupSchedule, err error) {
	log.Debug().Msg("updateBackupScheduleDB()")

	b, err = db.BackupSchedule.
		UpdateOneID(id).
		SetChoice(backupSchedule.Choice).
		SetCronSyntax(backupSchedule.CronSyntax).
		SetDay(backupSchedule.Day).
		SetHour(backupSchedule.Hour).
		SetEveryDay(backupSchedule.EveryDay).
		SetEveryWeek(backupSchedule.EveryWeek).
		SetEveryMonth(backupSchedule.EveryMonth).
		SetEveryMonday(backupSchedule.EveryMonday).
		SetEveryTuesday(backupSchedule.EveryTuesday).
		SetEveryWednesday(backupSchedule.EveryWednesday).
		SetEveryThursday(backupSchedule.EveryThursday).
		SetEveryFriday(backupSchedule.EveryFriday).
		SetEverySaturday(backupSchedule.EverySaturday).
		SetEverySunday(backupSchedule.EverySunday).
		SetInterval(backupSchedule.Interval).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}
	return b, err
}

func UpdateCronID(backupSchedule ent.BackupSchedule, newCronID int) (err error) {
	log.Debug().Msg("updateCronID()")

	ctx := context.Background()
	db := Connect()
	defer db.Close()

	_, err = db.BackupSchedule.UpdateOneID(backupSchedule.ID).SetCronID(newCronID).Save(ctx)
	if err != nil {
		log.Err(err).Msg("Error updating cronID")
	}

	return err
}

func UpdateBackupCronNextRun(backupSchedule ent.BackupSchedule, nextRun string) (err error) {
	log.Debug().Msg("UpdateBackupCronNextRun()")

	ctx := context.Background()
	db := Connect()
	defer db.Close()

	_, err = db.BackupSchedule.
		UpdateOneID(backupSchedule.ID).
		SetNextRun(nextRun).
		Save(ctx)
	if err != nil {
		log.Err(err).Msg("Error updating CronRuns")
	}

	return err
}

func UpdateBackupCronLastRun(backupSchedule ent.BackupSchedule, lastRun string) (err error) {
	log.Debug().Msg("UpdateBackupCronLastRun()")

	ctx := context.Background()
	db := Connect()
	defer db.Close()

	_, err = db.BackupSchedule.
		UpdateOneID(backupSchedule.ID).
		SetLastRun(lastRun).
		Save(ctx)
	if err != nil {
		log.Err(err).Msg("Error updating CronRuns")
	}

	return err
}

func QueryActiveRWithBackupSchedule() (r []*ent.Repository, err error) {
	ctx := context.Background()

	db := Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	r, err = db.Repository.Query().WithBackupSchedule().WithPruneSchedule().Where(repository.Active(true)).All(ctx)
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	}
	return r, err
}

func QueryPruneActiveRWithBackupSchedule() (r []*ent.Repository, err error) {
	ctx := context.Background()

	db := Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	r, err = db.Repository.Query().WithPruneSchedule().Where(repository.PruneActive(true)).All(ctx)
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	}
	return r, err
}

func QueryAllRWithBackupSchedule() (r []*ent.Repository, err error) {
	ctx := context.Background()

	db := Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	r, err = db.Repository.Query().WithBackupSchedule().WithPruneSchedule().All(ctx)
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	}
	return r, err
}
