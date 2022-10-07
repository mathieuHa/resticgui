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

// CreateLocalProvider godoc
// @Summary create backup local provider
// @Schemes
// @Description create a localprovider
// @Accept json
// @Param id path int true "Repository ID"
// @Param localprovider body ent.LocalProvider true "Create localprovider"
// @Produce json
// @Success 201 {object} ent.LocalProvider
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/localprovider [post]
func CreateLocalProvider(c *gin.Context) {
	log.Debug().Msg("CreateLocalProvider()")

	ctx := context.Background()
	var p ent.LocalProvider

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

	// create localprovider
	log.Debug().Msg("create localprovider")
	localProvider, err := database.CreateLocalProviderDB(p, rID, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, localProvider)
}

// UpdateLocalProvider godoc
// @Summary update local provider
// @Schemes
// @Description update a localprovider
// @Accept json
// @Param id path int true "Repository ID"
// @Param localprovider body ent.LocalProvider true "Update localprovider"
// @Produce json
// @Success 200 {object} ent.LocalProvider
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/localprovider [patch]
func UpdateLocalProvider(c *gin.Context) {
	log.Debug().Msg("UpdateLocalProvider()")

	ctx := context.Background()
	var p ent.LocalProvider
	var lp *ent.LocalProvider

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

	if r.Edges.LocalProvider == nil || r.Edges.LocalProvider.ID == 0 {
		lp, err = database.CreateLocalProviderDB(p, rID, ctx)
	} else {
		lp, err = database.UpdateLocalProviderDB(p, r.Edges.LocalProvider.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	updateRepositoryRunning(r, rID)

	c.IndentedJSON(http.StatusOK, lp)
}
