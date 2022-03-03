package API

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/repository"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Utils"
)

const entity = "repository "
const errNotFound = entity + "not found"
const errIDNotValid = "id not valid"
const errNotDeleted = entity + "not deleted"
const errNotBind = entity + "bind error"
const errUpdate = entity + "update error"
const errCreate = entity + "creation failed"
const errQuery = entity + "query failed"

// QueryRepositories godoc
// @Summary get all repositories
// @Schemes
// @Description returns all respositories
// @Accept json
// @Produce json
// @Success 200 {array} ent.Repository
// @Router /repositories [get]
func QueryRepositories(c *gin.Context) {

	db := Database.Connect()
	defer db.Close()

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
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	db := Database.Connect()
	defer db.Close()

	repository, err := db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		WithPaths().
		Where(repository.ID(newId)).
		Only(context.Background())

	if err != nil {
		log.Err(err).Msgf("error: %v", errNotFound)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errQuery, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, repository)
}

// GetRepository godoc
// @Summary create repository
// @Schemes
// @Description create a repository
// @Accept json
// @Param repository body ent.Repository true "Add account"
// @Produce json
// @Success 201 {object} ent.Repository
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories [post]
func CreateRepository(c *gin.Context) {
	ctx := context.Background()
	var repo ent.Repository

	if err := c.BindJSON(&repo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	db := Database.Connect()
	defer db.Close()

	if repo.Provider == "" {
		repo.Provider = "localProvider"
	}
	// create repository
	newRepository, err := db.Repository.
		Create().
		SetName(repo.Name).
		SetActive(repo.Active).
		SetPassword(repo.Password).
		SetProvider(repo.Provider).
		SetMaxSize(repo.MaxSize).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	// create backupSchedule
	if repo.Edges.BackupSchedule != nil {
		createBackupSchedule(db, *newRepository.Edges.BackupSchedule, newRepository.ID, ctx)
		// if err != nil {
		// 	log.Err(err).Msgf("error: %v", errQuery)
		// 	c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		// 	return
		// }
	}

	// create backupPolicy
	if repo.Edges.BackupPolicy != nil {
		createBackupPolicy(db, *newRepository.Edges.BackupPolicy, newRepository.ID, ctx)

		// if err != nil {
		// 	log.Err(err).Msgf("error: %v", errQuery)
		// 	c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
		// 	return
		// }
	}

	// create AWSS3Provider
	if repo.Edges.AWSS3Provider != nil {
		createAWSS3Provider(db, *newRepository.Edges.AWSS3Provider, newRepository.ID, ctx)
	}

	// create B2Provider
	if repo.Edges.B2Provider != nil {
		createB2Provider(db, *newRepository.Edges.B2Provider, newRepository.ID, ctx)
	}

	// create LocalProvider
	if repo.Edges.LocalProvider != nil {
		createLocalProvider(db, *newRepository.Edges.LocalProvider, newRepository.ID, ctx)
	}

	// create MinioS3Provider
	if repo.Edges.MinioS3Provider != nil {
		createMinioS3Provider(db, *newRepository.Edges.MinioS3Provider, newRepository.ID, ctx)
	}

	// Creates paths
	if repo.Edges.Paths != nil && len(repo.Edges.Paths) > 0 {
		for _, p := range repo.Edges.Paths {
			path, err := db.Path.
				Create().
				SetAbsPath(p.AbsPath).
				SetIsDir(p.IsDir).
				SetExists(p.Exists).
				SetIncluded(p.Included).
				SetName(p.Name).
				SetRelPath(p.RelPath).
				Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errQuery)
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				return
			}
			_, err = db.Repository.
				UpdateOneID(newRepository.ID).AddPaths(path).Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errQuery)
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				return
			}
		}
	}

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	repository, err := db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		WithPaths().
		Where(repository.ID(newRepository.ID)).
		Only(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errQuery, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, repository)
}

// GetRepository godoc
// @Summary edit repository
// @Schemes
// @Description edit a repository by ID
// @Accept json
// @Produce json
// @Param id path int true "Repository ID"
// @Param repository body ent.Repository true "Edit account"
// @Success 200 {object} ent.Repository
// @Failure 400 {string} errBindRepository
// @Failure 404 {string} errIDNotValid
// @Router /repositories/{id} [patch]
func EditRepository(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)
	var repo ent.Repository

	if err := c.BindJSON(&repo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	db := Database.Connect()
	defer db.Close()

	// Update repository
	_, err := db.Repository.
		UpdateOneID(newId).
		SetName(repo.Name).
		SetMaxSize(repo.MaxSize).
		SetActive(repo.Active).
		SetProvider(repo.Provider).
		Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: repository %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	// Update backup schedule
	if repo.Edges.BackupSchedule != nil {
		if repo.Edges.BackupSchedule.ID == 0 {
			log.Debug().Msgf("id = %v", repo.Edges.BackupSchedule.ID)
			createBackupSchedule(db, *repo.Edges.BackupSchedule, newId, ctx)
		} else {
			updateBackupSchedule(db, *repo.Edges.BackupSchedule, ctx)
		}
	}

	// Update backupPolicy
	if repo.Edges.BackupPolicy != nil {
		if repo.Edges.BackupPolicy.ID == 0 {
			createBackupPolicy(db, *repo.Edges.BackupPolicy, newId, ctx)
		} else {
			updateBackupPolicy(db, *repo.Edges.BackupPolicy, ctx)
		}
	}

	// Update AWSS3Provider
	if repo.Edges.AWSS3Provider != nil {

		if repo.Edges.AWSS3Provider.ID == 0 {
			createAWSS3Provider(db, *repo.Edges.AWSS3Provider, newId, ctx)
		} else {
			updateAWSS3Provider(db, *repo.Edges.AWSS3Provider, ctx)
		}

	}

	// Update B2Provider
	if repo.Edges.B2Provider != nil {
		if repo.Edges.B2Provider != nil {
			createB2Provider(db, *repo.Edges.B2Provider, newId, ctx)
		} else {
			updateB2Provider(db, *repo.Edges.B2Provider, ctx)
		}
	}

	// update LocalProvider
	if repo.Edges.LocalProvider != nil {
		if repo.Edges.LocalProvider != nil {
			createLocalProvider(db, *repo.Edges.LocalProvider, newId, ctx)
		} else {
			updateLocalProvider(db, *repo.Edges.LocalProvider, ctx)
		}
	}

	// update MinioS3Provider
	if repo.Edges.MinioS3Provider != nil {
		if repo.Edges.MinioS3Provider != nil {
			createMinioS3Provider(db, *repo.Edges.MinioS3Provider, newId, ctx)
		} else {
			updateMinioS3Provider(db, *repo.Edges.MinioS3Provider, ctx)
		}
	}

	// Remove old path edges
	_, err = db.Repository.
		UpdateOneID(newId).RemovePaths(db.Repository.Query().QueryPaths().AllX(ctx)...).Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	// Update path edges
	if repo.Edges.Paths != nil && len(repo.Edges.Paths) > 0 {
		for _, p := range repo.Edges.Paths {
			// ID has not been provided if p.ID == 0 -> new path
			path, err := db.Path.
				Create().
				SetAbsPath(p.AbsPath).
				SetIsDir(p.IsDir).
				SetExists(p.Exists).
				SetIncluded(p.Included).
				SetName(p.Name).
				SetRelPath(p.RelPath).
				Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errQuery)
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				return
			}
			_, err = db.Repository.
				UpdateOneID(newId).AddPaths(path).Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errQuery)
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				return
			}
		}
	}

	// Fetch repository after update
	repository, err := db.Repository.
		Query().
		WithBackupSchedule().
		WithBackupPolicy().
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		WithPaths().
		Where(repository.ID(newId)).
		First(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errQuery, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, repository)
}

// RemoveRepository godoc
// @Summary delete repository
// @Schemes
// @Description delete a repository by ID
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 204
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} errNotDeleted
// @Router /repositories/{id} [delete]
func RemoveRepository(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)
	var repo ent.Repository

	if err := c.BindJSON(&repo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	db := Database.Connect()
	defer db.Close()

	// Delete repository
	err := db.Repository.
		DeleteOneID(newId).
		Exec(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errNotDeleted, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
