package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mathieuHa/resticgui/ent"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/rs/zerolog/log"
)

// createAWSS3Provider godoc
func createAWSS3Provider(c *gin.Context, r *ent.Repository, p ent.AWSS3Provider) {
	log.Debug().Msg("CreateAWSS3Provider()")

	ctx := context.Background()

	// create s3Provider
	log.Debug().Msg("create s3Provider")
	s3Provider, err := database.CreateAWSS3ProviderDB(p, r.ID, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, s3Provider)
}

// updateAWSS3Provider godoc
func updateAWSS3Provider(c *gin.Context, r *ent.Repository, p ent.AWSS3Provider) {
	log.Debug().Msg("updateAWSS3Provider()")

	ctx := context.Background()
	var s3Provider *ent.AWSS3Provider
	var err error

	if r.Edges.AWSS3Provider == nil || r.Edges.AWSS3Provider.ID == 0 {
		s3Provider, err = database.CreateAWSS3ProviderDB(p, r.ID, ctx)
	} else {
		s3Provider, err = database.UpdateAWSS3ProviderDB(p, r.Edges.AWSS3Provider.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, s3Provider)
}
