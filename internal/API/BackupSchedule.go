package API

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"net/http"
	"time"
)

func QueryBackupSchedule(c *gin.Context) {
	var schedules []Models.BackupSchedule

	db := Database.Connect()

	db.Find(&schedules)

	c.IndentedJSON(http.StatusOK, schedules)
}

func GetBackupSchedule(c *gin.Context) {
	id := c.Param("id")
	if !Utils.IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

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
	log.Debug().Msg("UpdateBackup")
	id := c.Param("id") // id of the repository
	if !Utils.IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var repository Models.Repository

	db := Database.Connect()

	result := db.Preload("BackupSchedule").Find(&repository, "id = ?", id)
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

	if validSchedule(repository.BackupSchedule) == false {
		// we check the schedule can be parsed
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "schedule not valid"})
		log.Warn().Msg("schedule not valid")
		return
	}
	if repository.Active == true {
		// we add the scheduled backup
		resultAdd := addScheduledBackup(&repository)
		if resultAdd == false {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "start cron error"})
			log.Warn().Msg("start cron error")
			return
		}
	}

	db.Save(&repository)

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
		return schedule.Interval
	} else if schedule.Choice == "regular" {
		hour, _ := time.Parse("06h15m04s", schedule.Hour)
		day, _ := time.Parse("01", schedule.Day)
		log.Debug().Msgf("%v", hour)
		if schedule.EveryDay {
			return fmt.Sprintf("%d %d * * *", hour.Minute(), hour.Hour())
		} else if schedule.EveryWeek {
			return fmt.Sprintf("%d %d * * %d", hour.Minute(), hour.Hour(), day.Weekday())
		} else if schedule.EveryMonth {
			return fmt.Sprintf("%d %d %d * *", hour.Minute(), hour.Hour(), day.Day())
		}
	} else if schedule.Choice == "weekly" {
		// check weekly
		hour, _ := time.Parse("06h15m04s", schedule.Hour)
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
			day = day[:1]
		}
		return fmt.Sprintf("%d %d * * %v", hour.Minute(), hour.Hour(), day)
	}
	return ""
}

func addScheduledBackup(repository *Models.Repository) bool {
	id := repository.ID
	// we check if the cron has been initialized
	if cr != nil {

		// First we remove the previous job launched for this repository
		if repository.BackupSchedule.CronID >= 0 {
			cr.Remove(repository.BackupSchedule.CronID)
		}
		// Then we had the new job to be run and register it
		var err error
		repository.BackupSchedule.CronID, err = cr.AddFunc("@every 5s", func() { log.Debug().Msgf("Every hour thirty %v", id) })
		if err != nil {
			log.Error().Err(err)
			return false
		}
		return true
	} else {
		return false
	}

}

func removeScheduledBackup(repository Models.Repository) {
	// Here we need to
}
