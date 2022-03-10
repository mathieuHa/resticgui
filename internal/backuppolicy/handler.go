package backuppolicy

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"gitlab.com/resticgui/resticgui/internal/repository"
)

// CreateBackupPolicy godoc
// @Summary create backup schedule
// @Schemes
// @Description create a backuppolicy
// @Accept json
// @Param id path int true "Repository ID"
// @Param backupschedule body ent.BackupSchedule true "Create backuppolicy"
// @Produce json
// @Success 201 {object} ent.BackupSchedule
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/backuppolicy [post]
func CreateBackupPolicy(c *gin.Context) {
	log.Debug().Msg("CreateBackupPolicy()")

	ctx := context.Background()
	var b ent.BackupPolicy

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
	log.Debug().Msg("create backupPolicy")
	backupPolicy, err := CreateBackupPolicyDB(b, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, backupPolicy)
}

// UpdateBackupPolicy godoc
// @Summary update backup policy
// @Schemes
// @Description update a backuppolicy
// @Accept json
// @Param id path int true "Repository ID"
// @Param backuppolicy body ent.BackupPolicy true "Update backuppolicy"
// @Produce json
// @Success 200 {object} ent.BackupPolicy
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/backuppolicy [patch]
func UpdateBackupPolicy(c *gin.Context) {
	log.Debug().Msg("UpdateBackupPolicy()")

	ctx := context.Background()
	var b ent.BackupPolicy
	var backupPolicy *ent.BackupPolicy

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
		backupPolicy, err = CreateBackupPolicyDB(b, newId, ctx)
	} else {
		backupPolicy, err = UpdateBackupPolicyDB(b, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, backupPolicy)
}
