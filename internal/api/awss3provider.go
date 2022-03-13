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
)

// CreateAWSS3Provider godoc
// @Summary create AWSS3 provider
// @Schemes
// @Description create a AWSS3Provider
// @Accept json
// @Param id path int true "Repository ID"
// @Param s3provider body ent.AWSS3Provider true "Create AWSS3Provider"
// @Produce json
// @Success 201 {object} ent.AWSS3Provider
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/s3provider [post]
func CreateAWSS3Provider(c *gin.Context) {
	log.Debug().Msg("CreateAWSS3Provider()")

	ctx := context.Background()
	var s3 ent.AWSS3Provider
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&s3); err != nil {
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

	// create s3Provider
	log.Debug().Msg("create s3Provider")
	s3Provider, err := database.CreateAWSS3ProviderDB(s3, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, s3Provider)
}

// UpdateAWSS3Provider godoc
// @Summary update AWSS3 Provider
// @Schemes
// @Description update a s3provider
// @Accept json
// @Param id path int true "Repository ID"
// @Param s3provider body ent.AWSS3Provider true "Update s3provider"
// @Produce json
// @Success 200 {object} ent.AWSS3Provider
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/s3provider [patch]
func UpdateAWSS3Provider(c *gin.Context) {
	log.Debug().Msg("UpdateAWSS3Provider()")

	ctx := context.Background()
	var s3 ent.AWSS3Provider
	var s3Provider *ent.AWSS3Provider

	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&s3); err != nil {
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

	if r.Edges.AWSS3Provider == nil || r.Edges.AWSS3Provider.ID == 0 {
		s3Provider, err = database.CreateAWSS3ProviderDB(s3, newId, ctx)
	} else {
		s3Provider, err = database.UpdateAWSS3ProviderDB(s3, r.Edges.AWSS3Provider.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, s3Provider)
}
