package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/database"
	Models "gitlab.com/resticgui/resticgui/internal/models"
	"gitlab.com/resticgui/resticgui/internal/utils"
)

// CreateProvider godoc
// @Summary create provider
// @Schemes
// @Description create a provider
// @Accept json
// @Param id path int true "Repository ID"
// @Produce json
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/providers [post]
func CreateProvider(c *gin.Context) {
	log.Debug().Msg("CreateProvider()")

	ctx := context.Background()

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := checkRepositoryByID(rID, ctx, c); err != nil {
		return
	}
	r, _ := database.GetRepositoryByIDDB(rID, ctx)

	var p Models.Provider

	if err := c.BindJSON(&p); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": errNotBind})
		log.Info().Err(err).Msg(errNotBind)
		return
	}
	switch p.Provider {
	case "localProvider":
		createLocalProvider(c, r, p.LocalProvider)
	case "MinioS3Provider":
		createMinioS3Provider(c, r, p.MinioS3Provider)
	case "AWSS3Provider":
		createAWSS3Provider(c, r, p.AWSS3Provider)
	case "B2Provider":
		createB2Provider(c, r, p.B2Provider)
	default:
		err = errors.New("provider does not exist")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	_, err = database.SetRepositoryProviderDB(rID, p.Provider, ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error setting provider for repository")
	}
}

// UpdateProvider godoc
// @Summary update provider
// @Schemes
// @Description update a provider
// @Accept json
// @Param id path int true "Repository ID"
// @Produce json
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/provider [patch]
func UpdateProvider(c *gin.Context) {
	log.Debug().Msg("UpdateProvider()")

	ctx := context.Background()

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	if err := checkRepositoryByID(rID, ctx, c); err != nil {
		return
	}

	r, _ := database.GetRepositoryByIDDB(rID, ctx)

	var p Models.Provider

	if err := c.BindJSON(&p); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": errNotBind})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	switch p.Provider {
	case "localProvider":
		updateLocalProvider(c, r, p.LocalProvider)
	case "MinioS3Provider":
		updateMinioS3Provider(c, r, p.MinioS3Provider)
	case "AWSS3Provider":
		updateAWSS3Provider(c, r, p.AWSS3Provider)
	case "B2Provider":
		updateB2Provider(c, r, p.B2Provider)
	default:
		err = errors.New("provider does not exist")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	_, err = database.SetRepositoryProviderDB(rID, p.Provider, ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error setting provider for repository")
	}
}
