package API

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/repository"
	"gitlab.com/resticgui/resticgui/internal/Database"
)

// func QueryBackupSchedule(c *gin.Context) {
// 	var schedules []ent.BackupSchedule

// 	db := Database.Connect()

// 	db.Find(&schedules)

// 	c.IndentedJSON(http.StatusOK, schedules)
// }

// func GetBackupSchedule(c *gin.Context) {
// 	id := c.Param("id")

// 	var schedule ent.BackupSchedule

// 	db := Database.Connect()

// 	result := db.Find(&schedule, "id = ?", id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
// 		return
// 	}

// 	c.IndentedJSON(http.StatusOK, schedule)
// }

// func UpdateBackupSchedule(c *gin.Context) {
// 	id := c.Param("id") // id of the repository

// 	var repository ent.Repository

// 	db := Database.Connect()

// 	result := db.Preload("BackupSchedule").Find(&repository, id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
// 		return
// 	}

// 	// Here we add repository schedule
// 	//var schedule ent.BackupSchedule
// 	if err := c.BindJSON(&repository.BackupSchedule); err != nil {
// 		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "bind error"})
// 		log.Info().Err(err).Msg("Bind Error schedule")
// 		return
// 	}

// 	if !validSchedule(repository.BackupSchedule) {
// 		// we check the schedule can be parsed
// 		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "schedule not valid"})
// 		log.Warn().Msg("schedule not valid")
// 		return
// 	}
// 	if repository.Active {
// 		// we add the scheduled backup
// 		resultAdd := addScheduledBackup(&repository)
// 		if !resultAdd {
// 			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "start cron error"})
// 			log.Warn().Msg("start cron error")
// 			return
// 		}
// 	}

// 	db.Model(&repository).Association("BackupScheduled").Replace(&repository.BackupSchedule)

// 	// Check if the file exist ?

// 	c.IndentedJSON(http.StatusOK, repository.BackupSchedule)
// }

func validSchedule(schedule ent.BackupSchedule) bool {
	if schedule.Choice == "cron" {
		// check cron
	} else if schedule.Choice == "interval" {
		// check interval
	} else if schedule.Choice == "regular" {
		// check regular
	} else if schedule.Choice == "weekly" {
		// check weekly
	}
	return true
}

// This returns the schedule in string format that can be used by cron
func getScheduleStringSyntax(schedule ent.BackupSchedule) string {
	if schedule.Choice == "cron" {
		return schedule.CronSyntax
	} else if schedule.Choice == "interval" {
		return fmt.Sprintf("@every %v", schedule.Interval)
	} else if schedule.Choice == "regular" {
		hour, _ := time.Parse("15:04", schedule.Hour)
		log.Debug().Msgf("hour = %v", hour)
		if schedule.EveryDay {
			return fmt.Sprintf("%d %d * * *", hour.Minute(), hour.Hour())
		} else if schedule.EveryWeek {
			return fmt.Sprintf("%d %d * * %v", hour.Minute(), hour.Hour(), schedule.Day)
		} else if schedule.EveryMonth {
			return fmt.Sprintf("%d %d %v * *", hour.Minute(), hour.Hour(), schedule.Day)
		}
	} else if schedule.Choice == "weekly" {
		// check weekly
		hour, err := time.Parse("15:04", schedule.Hour)
		log.Debug().Err(err).Msgf("hour = %v", hour)
		day := ""
		if schedule.EveryMonday {
			day = fmt.Sprintf("%v1,", day)
		}
		if schedule.EveryTuesday {
			day = fmt.Sprintf("%v2,", day)
		}
		if schedule.EveryWednesday {
			day = fmt.Sprintf("%v3,", day)
		}
		if schedule.EveryThursday {
			day = fmt.Sprintf("%v4,", day)
		}
		if schedule.EveryFriday {
			day = fmt.Sprintf("%v5,", day)
		}
		if schedule.EverySaturday {
			day = fmt.Sprintf("%v6,", day)
		}
		if schedule.EverySunday {
			day = fmt.Sprintf("%v0,", day)
		}
		if len(day) > 0 {
			log.Debug().Msgf("DAY = %v", day)
			day = day[:len(day)-1]
			log.Debug().Msgf("DAY = %v", day)

		}
		return fmt.Sprintf("%d %d * * %v", hour.Minute(), hour.Hour(), day)
	}
	return ""
}

func addScheduledBackup(repository *ent.Repository) bool {
	// we check if the cron has been initialized
	if cr != nil && repository.Valid {

		// First we remove the previous job launched for this repository
		if *repository.Edges.BackupSchedule.CronID >= 0 {
			cr.Remove(cron.EntryID(*repository.Edges.BackupSchedule.CronID))
		}
		// Then we had the new job to be run and register it
		var err error
		var cronID cron.EntryID
		timing := getScheduleStringSyntax(*repository.Edges.BackupSchedule)
		log.Debug().Msgf("timing = %v", timing)
		cronID, err = cr.AddFunc(timing, func() {
			log.Info().Msgf("JOB HAS RUN %v", repository.ID)
		})
		if err != nil {
			log.Error().Err(err)
			return false
		}
		db := Database.Connect()
		defer db.Close()
		*repository.Edges.BackupSchedule.CronID = int(cronID)
		err = updateBackupSchedule(db, repository.Edges.BackupSchedule, context.Background())
		if err != nil {
			log.Err(err).Msgf("error: repository %v", errUpdate)
		}
		return true
	} else {
		return false
	}

}

func CreateAllScheduledBackup() {
	// var repositories []ent.Repository

	db := Database.Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	repositories, err := db.Repository.Query().Where(repository.Active(true)).All(context.Background())
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	}
	for _, repo := range repositories {
		log.Debug().Msgf("start repos %v", repo.ID)
		addScheduledBackup(repo)
	}
}

func removeScheduledBackup(repository ent.Repository) {
	// Here we need to
	if cr != nil {
		log.Debug().Msg("Removing backups")
		// First we remove the previous job launched for this repository
		if repository.Edges.BackupSchedule != nil && *repository.Edges.BackupSchedule.CronID > 0 {
			id := cron.EntryID(*repository.Edges.BackupSchedule.CronID)
			cr.Remove(id)
		}
	}
}

func createBackupSchedule(db *ent.Client, backupSchedule *ent.BackupSchedule, repoID int, ctx context.Context) (err error) {
	log.Debug().Msg("createBackupSchedule()")
	_, err = db.BackupSchedule.
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
		return err
	} else {
		return nil
	}
}

func updateBackupSchedule(db *ent.Client, backupSchedule *ent.BackupSchedule, ctx context.Context) (err error) {
	log.Debug().Msg("updateBackupSchedule()")
	_, err = db.BackupSchedule.
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
		return err
	} else {
		return nil
	}
}
