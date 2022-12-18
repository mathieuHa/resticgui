package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/database"
	"gitlab.com/resticgui/resticgui/internal/utils"
)

// CreatePruneSchedule godoc
// @Summary create prune schedule
// @Schemes
// @Description create a pruneschedule
// @Accept json
// @Param id path int true "Repository ID"
// @Param pruneschedule body ent.PruneSchedule true "Create pruneschedule"
// @Produce json
// @Success 201 {object} ent.PruneSchedule
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/pruneschedule [post]
func CreatePruneSchedule(c *gin.Context) {
	log.Debug().Msg("CreatePruneSchedule()")
	ctx := context.Background()
	var p ent.PruneSchedule

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := c.BindJSON(&p); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	if err := checkRepositoryByID(rID, ctx, c); err != nil {
		return
	}

	// create pruneSchedule
	log.Debug().Msg("create pruneSchedule")
	pruneSchedule, err := database.CreatePruneScheduleDB(p, rID, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, pruneSchedule)
}

// UpdatePruneSchedule godoc
// @Summary update backup schedule
// @Schemes
// @Description update a pruneschedule
// @Accept json
// @Param id path int true "Repository ID"
// @Param pruneschedule body ent.PruneSchedule true "Update pruneschedule"
// @Produce json
// @Success 200 {object} ent.PruneSchedule
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/pruneschedule [patch]
func UpdatePruneSchedule(c *gin.Context) {
	log.Debug().Msg("UpdatePruneSchedule()")

	ctx := context.Background()
	var p ent.PruneSchedule
	var pruneSchedule *ent.PruneSchedule

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := c.BindJSON(&p); err != nil {
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

	if r.Edges.PruneSchedule == nil || r.Edges.PruneSchedule.ID == 0 {
		pruneSchedule, err = database.CreatePruneScheduleDB(p, rID, ctx)
	} else {
		pruneSchedule, err = database.UpdatePruneScheduleDB(p, r.Edges.PruneSchedule.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	// updateRepositoryRunning(r, rID)

	c.IndentedJSON(http.StatusOK, pruneSchedule)
}
