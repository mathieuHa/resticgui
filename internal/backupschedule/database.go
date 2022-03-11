package backupschedule

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

func CreateBackupScheduleDB(backupSchedule ent.BackupSchedule, repoID int, ctx context.Context) (b *ent.BackupSchedule, err error) {
	log.Debug().Msg("CreateBackupScheduleDB()")

	db := Database.Connect()
	defer db.Close()

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
		SetRepositoryID(repoID).
		Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return b, err
}

func UpdateBackupScheduleDB(backupSchedule ent.BackupSchedule, ctx context.Context) (b *ent.BackupSchedule, err error) {
	log.Debug().Msg("updateBackupScheduleDB()")

	db := Database.Connect()
	defer db.Close()

	b, err = db.BackupSchedule.
		UpdateOneID(backupSchedule.ID).
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
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}
	return b, err
}

func updateCronID(backupSchedule ent.BackupSchedule, newCronID int) (err error) {
	log.Debug().Msg("updateCronID()")

	ctx := context.Background()
	db := Database.Connect()
	defer db.Close()

	_, err = db.BackupSchedule.UpdateOneID(backupSchedule.ID).SetCronID(newCronID).Save(ctx)
	if err != nil {
		log.Err(err).Msg("Error updating cronID")
	}

	return err
}

func logCommand(id int, operation string, stdout string, stderr string, err error, ctx context.Context) {
	db := Database.Connect()
	defer db.Close()

	var errStr string
	if err == nil {
		errStr = ""
	} else {
		errStr = err.Error()
		log.Err(err).Msg(operation)
	}
	_, errDB := db.Logs.
		Create().
		SetOperation(operation).
		SetStdout(stdout).
		SetStderr(stderr).
		SetError(errStr).
		SetRepositoryID(id).
		Save(ctx)

	if errDB != nil {
		log.Err(errDB)
	}
}
