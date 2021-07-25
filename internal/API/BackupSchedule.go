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

	var editedSchedule Models.BackupSchedule

	if err := c.BindJSON(&editedSchedule); err != nil {
		log.Info().Msg("Bind Error Schedule")
		return
	}
	// check on invalid data to do
	// errors.Is(result.Error, gorm.ErrInvalidData)

	db.Save(&editedSchedule)
	c.IndentedJSON(http.StatusOK, schedule)
}
