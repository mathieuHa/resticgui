package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mathieuHa/resticgui/ent"
	"github.com/mathieuHa/resticgui/internal/database"
	"github.com/rs/zerolog/log"
)

// createB2Provider godoc
func createB2Provider(c *gin.Context, r *ent.Repository, p ent.B2Provider) {
	log.Debug().Msg("createB2Provider()")

	ctx := context.Background()

	// create b2Provider
	log.Debug().Msg("create b2Provider")
	b2Provider, err := database.CreateB2ProviderDB(p, r.ID, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, b2Provider)
}

// updateB2Provider godoc
func updateB2Provider(c *gin.Context, r *ent.Repository, p ent.B2Provider) {
	log.Debug().Msg("updateB2Provider()")

	ctx := context.Background()
	var b2Provider *ent.B2Provider
	var err error

	if r.Edges.B2Provider == nil || r.Edges.B2Provider.ID == 0 {
		b2Provider, err = database.CreateB2ProviderDB(p, r.ID, ctx)
	} else {
		b2Provider, err = database.UpdateB2ProviderDB(p, r.Edges.B2Provider.ID, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, b2Provider)
}
