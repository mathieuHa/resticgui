package API

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"net/http"
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
		log.Info().Err(err).Msg("Bind Error schedule")
		return
	}

	validSchedule(repository.BackupSchedule)

	//err := db.Model(&repository).Association("Entities").Replace(entities)
	//if err != nil {
	//	c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error adding entities"})
	//	return
	//}

	//log.Debug().Msg(schedule.CronSyntax)
	//repository.BackupSchedule = schedule

	//db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&repository)

	db.Save(&repository)

	// Check if the file exist ?

	c.IndentedJSON(http.StatusOK, repository.BackupSchedule)
}

func validSchedule(schedule Models.BackupSchedule) {
    if (schedule.Choice == "cron") {
        // check cron
    } else if (schedule.Choice == "interval") {
        // check interval
    } else if (schedule.Choice == "regular") {
        // check regular
    } else if (schedule.Choice == "weekly") {
        // check weekly
    }

}