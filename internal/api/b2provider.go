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

// CreateB2Provider godoc
// @Summary create B2 provider
// @Schemes
// @Description create a B2Provider
// @Accept json
// @Param id path int true "Repository ID"
// @Param b2provider body ent.B2Provider true "Create B2Provider"
// @Produce json
// @Success 201 {object} ent.B2Provider
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/b2provider [post]
func CreateB2Provider(c *gin.Context) {
	log.Debug().Msg("CreateB2Provider()")

	ctx := context.Background()
	var b2 ent.B2Provider

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := c.BindJSON(&b2); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	if err := checkRepositoryByID(rID, ctx, c); err != nil {
		return
	}

	// create b2Provider
	log.Debug().Msg("create b2Provider")
	b2Provider, err := database.CreateB2ProviderDB(b2, rID, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, b2Provider)
}

// UpdateB2Provider godoc
// @Summary update B2 Provider
// @Schemes
// @Description update a b2provider
// @Accept json
// @Param id path int true "Repository ID"
// @Param b2provider body ent.B2Provider true "Update b2provider"
// @Produce json
// @Success 200 {object} ent.B2Provider
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/b2provider [patch]
func UpdateB2Provider(c *gin.Context) {
	log.Debug().Msg("UpdateB2Provider()")

	ctx := context.Background()
	var b2 ent.B2Provider
	var b2Provider *ent.B2Provider

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := c.BindJSON(&b2); err != nil {
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

	if r.Edges.B2Provider == nil || r.Edges.B2Provider.ID == 0 {
		b2Provider, err = database.CreateB2ProviderDB(b2, rID, ctx)
	} else {
		b2Provider, err = database.UpdateB2ProviderDB(b2, r.Edges.B2Provider.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	updateRepositoryRunning(rID)

	c.IndentedJSON(http.StatusOK, b2Provider)
}
