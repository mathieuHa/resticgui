package repository

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// QueryRepositories godoc
// @Summary get all repositories
// @Schemes
// @Description returns all respositories
// @Accept json
// @Produce json
// @Success 200 {array} ent.Repository
// @Router /repositories [get]
func QueryRepositories(c *gin.Context) {

	repositories, err := db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		WithPaths().
		All(context.Background())

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errQuery, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, repositories)
}
