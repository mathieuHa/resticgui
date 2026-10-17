package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mathieuHa/resticgui/ent"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/mathieuHa/resticgui/internal/utils"
	"github.com/rs/zerolog/log"
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

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := c.BindJSON(&b); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	if err := checkRepositoryByID(rID, ctx, c); err != nil {
		return
	}

	// create backupSchedule
	log.Debug().Msg("create backupPolicy")
	backupPolicy, err := database.CreateBackupPolicyDB(b, rID, ctx)
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

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := c.BindJSON(&b); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	r, err := database.GetRepositoryByIDDB(rID, ctx)
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
		backupPolicy, err = database.CreateBackupPolicyDB(b, rID, ctx)
	} else {
		backupPolicy, err = database.UpdateBackupPolicyDB(b, r.Edges.BackupPolicy.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	updateRepositoryRunning(rID)

	c.IndentedJSON(http.StatusOK, backupPolicy)
}
