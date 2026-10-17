package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mathieuHa/resticgui/ent"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/rs/zerolog/log"
)

// createLocalProvider godoc
func createLocalProvider(c *gin.Context, r *ent.Repository, p ent.LocalProvider) {
	log.Debug().Msg("CreateLocalProvider()")

	ctx := context.Background()

	// create localprovider
	log.Debug().Msg("create localprovider")
	localProvider, err := database.CreateLocalProviderDB(p, r.ID, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, localProvider)
}

// updateLocalProvider godoc
func updateLocalProvider(c *gin.Context, r *ent.Repository, p ent.LocalProvider) {
	log.Debug().Msg("updateLocalProvider()")

	ctx := context.Background()
	var lp *ent.LocalProvider
	var err error

	if r.Edges.LocalProvider == nil || r.Edges.LocalProvider.ID == 0 {
		lp, err = database.CreateLocalProviderDB(p, r.ID, ctx)
	} else {
		lp, err = database.UpdateLocalProviderDB(p, r.Edges.LocalProvider.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	updateRepositoryRunning(r.ID)

	c.IndentedJSON(http.StatusOK, lp)
}
