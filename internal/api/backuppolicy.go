package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/database"
	"gitlab.com/resticgui/resticgui/internal/utils"
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

	id := c.Param("id")

	if !utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&b); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	newId, _ := strconv.Atoi(id)

	_, err := database.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
		} else if ent.IsConstraintError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errRConstraint})
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		}
		return
	}

	// create backupSchedule
	log.Debug().Msg("create backupPolicy")
	backupPolicy, err := database.CreateBackupPolicyDB(b, newId, ctx)
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

	if !utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&b); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	newId, _ := strconv.Atoi(id)

	r, err := database.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
		} else if ent.IsConstraintError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errRConstraint})
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		}
		return
	}

	if r.Edges.BackupPolicy == nil || r.Edges.BackupPolicy.ID == 0 {
		backupPolicy, err = database.CreateBackupPolicyDB(b, newId, ctx)
	} else {
		backupPolicy, err = database.UpdateBackupPolicyDB(b, r.Edges.BackupPolicy.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	updateRepositoryRunning(r, newId)

	c.IndentedJSON(http.StatusOK, backupPolicy)
}
