package repository

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Utils"
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
	ctx := context.Background()

	repositories, err := GetRepositoriesDB(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, repositories)
}

// GetRepository godoc
// @Summary get repository
// @Schemes
// @Description returns a respository filtered by his ID
// @Accept json
// @Produce json
// @Param id path int true "Repository ID"
// @Success 200 {object} ent.Repository
// @Failure 404 {string} errIDNotValid
// @Router /repositories/{id} [get]
func GetRepository(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	r, err := GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, r)
}

// CreateRepository godoc
// @Summary create repository
// @Schemes
// @Description create a repository
// @Accept json
// @Param repository body ent.Repository true "Create repository"
// @Produce json
// @Success 201 {object} ent.Repository
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories [post]
func CreateRepository(c *gin.Context) {
	log.Debug().Msg("CreateRepository()")
	ctx := context.Background()
	var repo ent.Repository

	if err := c.BindJSON(&repo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	r, err := CreateRepositoryDB(repo, ctx)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	// // create backupSchedule
	// if repo.Edges.BackupSchedule != nil {
	// 	log.Debug().Msg("create backupSchedule")
	// 	err = createBackupSchedule(db, repo.Edges.BackupSchedule, newRepository.ID, ctx)
	// 	if err != nil {
	// 		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
	// 		return
	// 	}
	// }

	// // create backupPolicy
	// if repo.Edges.BackupPolicy != nil {
	// 	log.Debug().Msg("create backupPolicy")
	// 	err = createBackupPolicy(db, repo.Edges.BackupPolicy, newRepository.ID, ctx)
	// 	if err != nil {
	// 		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
	// 		return
	// 	}
	// }

	// // create AWSS3Provider
	// log.Debug().Msg("create AWSS3Provider")
	// if repo.Edges.AWSS3Provider != nil {
	// 	err = createAWSS3Provider(db, repo.Edges.AWSS3Provider, newRepository.ID, ctx)
	// 	if err != nil {
	// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
	// 		return
	// 	}
	// }

	// // create B2Provider
	// log.Debug().Msg("create B2Provider")
	// if repo.Edges.B2Provider != nil {
	// 	err = createB2Provider(db, repo.Edges.B2Provider, newRepository.ID, ctx)
	// 	if err != nil {
	// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
	// 		return
	// 	}
	// }

	// // create LocalProvider
	// if repo.Edges.LocalProvider != nil {
	// 	createLocalProvider(db, repo.Edges.LocalProvider, newRepository.ID, ctx)
	// }

	// // create MinioS3Provider
	// if repo.Edges.MinioS3Provider != nil {
	// 	createMinioS3Provider(db, repo.Edges.MinioS3Provider, newRepository.ID, ctx)
	// }

	// // Creates paths
	// createPaths(db, repo.Edges.Paths, newRepository.ID, ctx)

	r.Valid, err = validRepository(*r)
	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
	}

	// if r.Active && r.Valid {
	// 	addScheduledBackup(r)
	// }

	c.IndentedJSON(http.StatusCreated, r)
}
