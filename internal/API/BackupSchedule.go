package API

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
)

func QueryBackupSchedule(c *gin.Context) {
	var schedules []Models.BackupSchedule

	db := Database.Connect()

	db.Find(&schedules)

	c.IndentedJSON(http.StatusOK, schedules)
}

func GetBackupSchedule(c *gin.Context) {
	id := c.Param("id")

	var schedule Models.BackupSchedule

	db := Database.Connect()

	result := db.Find(&schedule, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, schedule)
}

func UpdateBackupSchedule(c *gin.Context) {
	id := c.Param("id") // id of the repository

	var repository Models.Repository

	db := Database.Connect()

	result := db.Preload("BackupSchedule").Find(&repository, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	// Here we add repository schedule
	//var schedule Models.BackupSchedule
	if err := c.BindJSON(&repository.BackupSchedule); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "bind error"})
		log.Info().Err(err).Msg("Bind Error schedule")
		return
	}

	if !validSchedule(repository.BackupSchedule) {
		// we check the schedule can be parsed
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "schedule not valid"})
		log.Warn().Msg("schedule not valid")
		return
	}
	if repository.Active {
		// we add the scheduled backup
		resultAdd := addScheduledBackup(&repository)
		if !resultAdd {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "start cron error"})
			log.Warn().Msg("start cron error")
			return
		}
	}

	db.Model(&repository).Association("BackupScheduled").Replace(&repository.BackupSchedule)

	// Check if the file exist ?

	c.IndentedJSON(http.StatusOK, repository.BackupSchedule)
}

func validSchedule(schedule Models.BackupSchedule) bool {
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
func getScheduleStringSyntax(schedule Models.BackupSchedule) string {
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

func addScheduledBackup(repository *Models.Repository) bool {
	// we check if the cron has been initialized
	if cr != nil {

		// First we remove the previous job launched for this repository
		if repository.BackupSchedule.CronID >= 0 {
			cr.Remove(repository.BackupSchedule.CronID)
		}
		// Then we had the new job to be run and register it
		var err error
		timing := getScheduleStringSyntax(repository.BackupSchedule)
		log.Debug().Msgf("timing = %v", timing)
		repository.BackupSchedule.CronID, err = cr.AddFunc(timing, func() { log.Debug().Msgf("JOB HAS RUN") })
		if err != nil {
			log.Error().Err(err)
			return false
		}
		return true
	} else {
		return false
	}

}

func CreateAllScheduledBackup() {
	var repositories []Models.Repository

	db := Database.Connect()

	db.Where("active = ?", true).Preload("BackupSchedule").Find(&repositories)

	for _, repo := range repositories {
		log.Debug().Msgf("start repos %v", repo.ID)
		addScheduledBackup(&repo)
	}
}

func removeScheduledBackup(repository Models.Repository) {
	// Here we need to
}
