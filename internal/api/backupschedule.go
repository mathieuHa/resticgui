package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"gitlab.com/resticgui/resticgui/internal/database"
	"gitlab.com/resticgui/resticgui/internal/repository"
)

// CreateBackupSchedule godoc
// @Summary create backup schedule
// @Schemes
// @Description create a backupschedule
// @Accept json
// @Param id path int true "Repository ID"
// @Param backupschedule body ent.BackupSchedule true "Create backupschedule"
// @Produce json
// @Success 201 {object} ent.BackupSchedule
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/backupschedule [post]
func CreateBackupSchedule(c *gin.Context) {
	log.Debug().Msg("CreateBackupSchedule()")
	ctx := context.Background()
	var b ent.BackupSchedule

	if err := c.BindJSON(&b); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	_, err := repository.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// create backupSchedule
	log.Debug().Msg("create backupSchedule")
	backupSchedule, err := database.CreateBackupScheduleDB(b, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, backupSchedule)
}

// UpdateBackupSchedule godoc
// @Summary update backup schedule
// @Schemes
// @Description update a backupschedule
// @Accept json
// @Param id path int true "Repository ID"
// @Param backupschedule body ent.BackupSchedule true "Update backupschedule"
// @Produce json
// @Success 200 {object} ent.BackupSchedule
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/backupschedule [patch]
func UpdateBackupSchedule(c *gin.Context) {
	log.Debug().Msg("UpdateBackupSchedule()")

	ctx := context.Background()
	var b ent.BackupSchedule
	var backupSchedule *ent.BackupSchedule

	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&b); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	newId, _ := strconv.Atoi(id)

	_, err := repository.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if b.ID == 0 && err != nil {
		backupSchedule, err = database.CreateBackupScheduleDB(b, newId, ctx)
	} else {
		backupSchedule, err = database.UpdateBackupScheduleDB(b, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, backupSchedule)
}
