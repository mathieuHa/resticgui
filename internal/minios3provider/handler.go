package minios3provider

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

// CreateMinioS3Provider godoc
// @Summary create backup local provider
// @Schemes
// @Description create a minios3provider
// @Accept json
// @Param id path int true "Repository ID"
// @Param minios3provider body ent.MinioS3Provider true "Create minios3provider"
// @Produce json
// @Success 201 {object} ent.MinioS3Provider
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/minios3provider [post]
func CreateMinioS3Provider(c *gin.Context) {
	log.Debug().Msg("CreateMinioS3Provider()")

	ctx := context.Background()
	var p ent.MinioS3Provider

	if err := c.BindJSON(&p); err != nil {
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

	// create minios3provider
	log.Debug().Msg("create minios3provider")
	backupPolicy, err := CreateMinioS3ProviderDB(p, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, backupPolicy)
}

// UpdateMinioS3Provider godoc
// @Summary update local provider
// @Schemes
// @Description update a minios3provider
// @Accept json
// @Param id path int true "Repository ID"
// @Param minios3provider body ent.MinioS3Provider true "Update minios3provider"
// @Produce json
// @Success 200 {object} ent.MinioS3Provider
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/minios3provider [patch]
func UpdateMinioS3Provider(c *gin.Context) {
	log.Debug().Msg("UpdateMinioS3Provider()")

	ctx := context.Background()
	var p ent.MinioS3Provider
	var lp *ent.MinioS3Provider

	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&p); err != nil {
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

	if p.ID == 0 && err != nil {
		lp, err = CreateMinioS3ProviderDB(p, newId, ctx)
	} else {
		lp, err = UpdateMinioS3ProviderDB(p, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, lp)
}
