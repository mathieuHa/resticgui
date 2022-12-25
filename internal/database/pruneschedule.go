package database

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/repository"
)

func CreatePruneScheduleDB(pruneSchedule ent.PruneSchedule, repoID int, ctx context.Context) (b *ent.PruneSchedule, err error) {
	log.Debug().Msg("CreatePruneScheduleDB()")

	b, err = db.PruneSchedule.
		Create().
		SetChoice(pruneSchedule.Choice).
		SetCronSyntax(pruneSchedule.CronSyntax).
		SetDay(pruneSchedule.Day).
		SetHour(pruneSchedule.Hour).
		SetEveryDay(pruneSchedule.EveryDay).
		SetEveryWeek(pruneSchedule.EveryWeek).
		SetEveryMonth(pruneSchedule.EveryMonth).
		SetEveryMonday(pruneSchedule.EveryMonday).
		SetEveryTuesday(pruneSchedule.EveryTuesday).
		SetEveryWednesday(pruneSchedule.EveryWednesday).
		SetEveryThursday(pruneSchedule.EveryThursday).
		SetEveryFriday(pruneSchedule.EveryFriday).
		SetEverySaturday(pruneSchedule.EverySaturday).
		SetEverySunday(pruneSchedule.EverySunday).
		SetInterval(pruneSchedule.Interval).
		SetRepositoryID(repoID).
		Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}

	return b, err
}

func UpdatePruneScheduleDB(pruneSchedule ent.PruneSchedule, id int, ctx context.Context) (b *ent.PruneSchedule, err error) {
	log.Debug().Msg("updatePruneScheduleDB()")

	b, err = db.PruneSchedule.
		UpdateOneID(id).
		SetChoice(pruneSchedule.Choice).
		SetCronSyntax(pruneSchedule.CronSyntax).
		SetDay(pruneSchedule.Day).
		SetHour(pruneSchedule.Hour).
		SetEveryDay(pruneSchedule.EveryDay).
		SetEveryWeek(pruneSchedule.EveryWeek).
		SetEveryMonth(pruneSchedule.EveryMonth).
		SetEveryMonday(pruneSchedule.EveryMonday).
		SetEveryTuesday(pruneSchedule.EveryTuesday).
		SetEveryWednesday(pruneSchedule.EveryWednesday).
		SetEveryThursday(pruneSchedule.EveryThursday).
		SetEveryFriday(pruneSchedule.EveryFriday).
		SetEverySaturday(pruneSchedule.EverySaturday).
		SetEverySunday(pruneSchedule.EverySunday).
		SetInterval(pruneSchedule.Interval).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
	}
	return b, err
}

func UpdatePruneCronID(schedule ent.PruneSchedule, newCronID int) (err error) {
	log.Debug().Msg("UpdatePruneCronID()")

	ctx := context.Background()
	db := Connect()
	defer db.Close()

	_, err = db.PruneSchedule.UpdateOneID(schedule.ID).SetCronID(newCronID).Save(ctx)
	if err != nil {
		log.Err(err).Msg("Error updating cronID")
	}

	return err
}

func UpdatePruneCronNextRun(schedule ent.PruneSchedule, nextRun string) (err error) {
	log.Debug().Msg("UpdatePruneCronNextRun()")

	ctx := context.Background()
	db := Connect()
	defer db.Close()

	_, err = db.PruneSchedule.
		UpdateOneID(schedule.ID).
		SetNextRun(nextRun).
		Save(ctx)
	if err != nil {
		log.Err(err).Msg("Error updating CronRuns")
	}

	return err
}

func UpdatePruneCronLastRun(schedule ent.PruneSchedule, lastRun string) (err error) {
	log.Debug().Msg("UpdatePruneCronLastRun()")

	ctx := context.Background()
	db := Connect()
	defer db.Close()

	log.Debug().Msgf("Last Run %s", lastRun)
	_, err = db.PruneSchedule.
		UpdateOneID(schedule.ID).
		SetLastRun(lastRun).
		Save(ctx)
	if err != nil {
		log.Err(err).Msg("Error updating CronRuns")
	}

	return err
}

func QueryActiveRWithPruneSchedule() (r []*ent.Repository, err error) {
	ctx := context.Background()

	db := Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	r, err = db.Repository.Query().WithPruneSchedule().Where(repository.Active(true)).All(ctx)
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	}
	return r, err
}

func QueryAllRWithPruneSchedule() (r []*ent.Repository, err error) {
	ctx := context.Background()

	db := Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	r, err = db.Repository.Query().WithPruneSchedule().All(ctx)
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	}
	return r, err
}
