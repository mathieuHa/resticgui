package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/database"
)

// createMinioS3Provider godoc
func createMinioS3Provider(c *gin.Context, r *ent.Repository, p ent.MinioS3Provider) {
	log.Debug().Msg("createMinioS3Provider()")

	ctx := context.Background()
	var err error

	// create minios3provider
	log.Debug().Msg("create minios3provider")
	ms3p, err := database.CreateMinioS3ProviderDB(p, r.ID, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, ms3p)
}

// updateMinioS3Provider godoc
func updateMinioS3Provider(c *gin.Context, r *ent.Repository, p ent.MinioS3Provider) {
	log.Debug().Msg("updateMinioS3Provider()")

	ctx := context.Background()
	var ms3p *ent.MinioS3Provider
	var err error

	if r.Edges.MinioS3Provider == nil || r.Edges.MinioS3Provider.ID == 0 {
		ms3p, err = database.CreateMinioS3ProviderDB(p, r.ID, ctx)
	} else {
		ms3p, err = database.UpdateMinioS3ProviderDB(p, r.Edges.MinioS3Provider.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	updateRepositoryRunning(r.ID)

	c.IndentedJSON(http.StatusOK, ms3p)
}
