package API

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/repository"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/commands"
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

func validSchedule(schedule *ent.BackupSchedule) (valid bool, err error) {
	log.Debug().Msg("validSchedule()")

	if schedule == nil {
		valid = false
		err = errors.New("schedule is nil")
		return valid, err
	}
	switch schedule.Choice {
	// check cron
	case "cron":
		return true, nil
		// check interval
	case "interval":
		return true, nil
		// check regular
	case "regular":
		return true, nil
		// check weekly
	case "weekly":
		return true, nil
	default:
		valid = false
		err = errors.New("provider does not exist")
	}
	return valid, err
}

// This returns the schedule in string format that can be used by cron
func getScheduleStringSyntax(schedule ent.BackupSchedule) string {
	log.Debug().Msg("getScheduleStringSyntax()")

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

func addScheduledBackup(repo *ent.Repository) bool {
	log.Debug().Msg("addScheduledBackup()")

	// we check if the cron has been initialized
	if cr != nil && repo.Valid {
		log.Debug().Msg("repository is valid")

		// First we remove the previous job launched for this repository
		log.Debug().Msg("remove previous job for the repository")
		if repo.Edges.BackupSchedule != nil {
			if repo.Edges.BackupSchedule.CronID != nil {
				log.Debug().Msg("in if")
				log.Debug().Msgf("cron id: %v", *repo.Edges.BackupSchedule.CronID)
				cr.Remove(cron.EntryID(*repo.Edges.BackupSchedule.CronID))
			} else {
				log.Debug().Msg("in else")
			}
			// Then we had the new job to be run and register it
			timing := getScheduleStringSyntax(*repo.Edges.BackupSchedule)
			log.Debug().Msgf("timing = %v", timing)
			cronID, err := cr.AddFunc(timing, func() {
				log.Info().Msgf("JOB HAS RUN %v", repo.ID)
				log.Info().Msgf("JOB HAS RUN %v", repo.Name)
				commands.GetResticVersion()
				db := Database.Connect()
				defer db.Close()
				// Check the validity of the provider and repository in general
				r, _ := db.Repository.
					Query().
					WithBackupSchedule().
					WithBackupPolicy().
					WithAWSS3Provider().
					WithB2Provider().
					WithLocalProvider().
					WithMinioS3Provider().
					WithPaths().
					Where(repository.ID(repo.ID)).
					First(context.Background())

				commands.InitRepository(*r)
			})
			if err != nil {
				log.Error().Err(err)
				return false
			}
			log.Info().Msgf("cronID: %v", int(cronID))

			err = updateCronID(*repo.Edges.BackupSchedule, int(cronID))
			if err != nil {
				log.Err(err).Msgf("error: repository %v", errUpdate)
			}
		}
		return true
	} else {
		log.Debug().Msg("repository is not valid")
		return false
	}

}

func CreateAllScheduledBackup() {
	log.Debug().Msg("CreateAllScheduledBackup()")

	db := Database.Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	repositories, err := db.Repository.Query().WithBackupSchedule().Where(repository.Active(true)).All(context.Background())
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	}
	for _, repo := range repositories {
		log.Debug().Msgf("start repos %v", repo.ID)
		addScheduledBackup(repo)
	}
}

func RemoveAllScheduledBackup() {
	log.Debug().Msg("RemoveAllScheduledBackup()")

	db := Database.Connect()
	defer db.Close()
	// Check the validity of the provider and repository in general
	repositories, err := db.Repository.Query().WithBackupSchedule().All(context.Background())
	if err != nil {
		log.Err(err).Msgf("error quering active backups")
	} else {
		for _, repo := range repositories {
			log.Debug().Msgf("start repos %v", repo.ID)
			removeScheduledBackup(repo)
		}
	}

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

func removeScheduledBackup(repository *ent.Repository) {
	log.Debug().Msg("removeScheduledBackup()")

	if repository == nil {
		return
	}
	// Here we need to
	if cr != nil {
		log.Debug().Msg("Removing backups")
		// First we remove the previous job launched for this repository
		if repository.Edges.BackupSchedule != nil && repository.Edges.BackupSchedule.CronID != nil {
			log.Debug().Msg("repository.Edges.BackupSchedule != nil && *repository.Edges.BackupSchedule.CronID > 0")

			id := cron.EntryID(*repository.Edges.BackupSchedule.CronID)
			cr.Remove(id)
			updateCronID(*repository.Edges.BackupSchedule, 0)
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
