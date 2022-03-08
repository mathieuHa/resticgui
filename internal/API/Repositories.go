package API

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/ent/logs"
	"gitlab.com/resticgui/resticgui/ent/repository"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Utils"
)

const entity = "repository "
const errNotFound = entity + "not found"
const errIDNotValid = "id not valid"
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
	log.Debug().Msg("CreateRepository()")
	ctx := context.Background()
	var repo ent.Repository

	if err := c.BindJSON(&repo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	log.Debug().Msg("DatabaseConnect()")
	db := Database.Connect()
	defer db.Close()

	if repo.Provider == "" {
		repo.Provider = "localProvider"
	}

	// create repository
	log.Debug().Msg("create repository")
	newRepository, err := db.Repository.
		Create().
		SetName(repo.Name).
		SetActive(repo.Active).
		SetRunning(false).
		SetPassword(repo.Password).
		SetProvider(repo.Provider).
		SetMaxSize(repo.MaxSize).
		Save(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errCreate)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	// create backupSchedule
	if repo.Edges.BackupSchedule != nil {
		log.Debug().Msg("create backupSchedule")
		err = createBackupSchedule(db, repo.Edges.BackupSchedule, newRepository.ID, ctx)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create backupPolicy
	if repo.Edges.BackupPolicy != nil {
		log.Debug().Msg("create backupPolicy")
		err = createBackupPolicy(db, repo.Edges.BackupPolicy, newRepository.ID, ctx)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create AWSS3Provider
	log.Debug().Msg("create AWSS3Provider")
	if repo.Edges.AWSS3Provider != nil {
		err = createAWSS3Provider(db, repo.Edges.AWSS3Provider, newRepository.ID, ctx)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create B2Provider
	log.Debug().Msg("create B2Provider")
	if repo.Edges.B2Provider != nil {
		err = createB2Provider(db, repo.Edges.B2Provider, newRepository.ID, ctx)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create LocalProvider
	if repo.Edges.LocalProvider != nil {
		createLocalProvider(db, repo.Edges.LocalProvider, newRepository.ID, ctx)
	}

	// create MinioS3Provider
	if repo.Edges.MinioS3Provider != nil {
		createMinioS3Provider(db, repo.Edges.MinioS3Provider, newRepository.ID, ctx)
	}

	// Creates paths
	createPaths(db, repo.Edges.Paths, newRepository.ID, ctx)

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

	repository.Valid, err = validRepository(*repository)
	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
	}

	if repository.Active && repository.Valid {
		addScheduledBackup(repository)
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
	r, err := db.Repository.
		UpdateOneID(newId).
		SetName(repo.Name).
		SetMaxSize(repo.MaxSize).
		SetActive(repo.Active).
		SetPassword(repo.Password).
		SetProvider(repo.Provider).
		Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: repository %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	// Update backup schedule
	if repo.Edges.BackupSchedule != nil {
		_, err := db.Repository.QueryBackupSchedule(r).FirstID(ctx)
		if repo.Edges.BackupSchedule.ID == 0 && err != nil {
			log.Debug().Msgf("id = %v", repo.Edges.BackupSchedule.ID)
			err = createBackupSchedule(db, repo.Edges.BackupSchedule, newId, ctx)
		} else {
			err = updateBackupSchedule(db, repo.Edges.BackupSchedule, ctx)
		}
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// Update backupPolicy
	if repo.Edges.BackupPolicy != nil {
		if repo.Edges.BackupPolicy.ID == 0 {
			err = createBackupPolicy(db, repo.Edges.BackupPolicy, newId, ctx)
		} else {
			err = updateBackupPolicy(db, repo.Edges.BackupPolicy, ctx)
		}
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// Update AWSS3Provider
	if repo.Edges.AWSS3Provider != nil {
		if repo.Edges.AWSS3Provider.ID == 0 {
			err = createAWSS3Provider(db, repo.Edges.AWSS3Provider, newId, ctx)
		} else {
			err = updateAWSS3Provider(db, repo.Edges.AWSS3Provider, ctx)
		}
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}

	}

	// Update B2Provider
	if repo.Edges.B2Provider != nil {
		if repo.Edges.B2Provider.ID == 0 {
			err = createB2Provider(db, repo.Edges.B2Provider, newId, ctx)
		} else {
			err = updateB2Provider(db, repo.Edges.B2Provider, ctx)
		}
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// update LocalProvider
	if repo.Edges.LocalProvider != nil {
		_, err := db.Repository.QueryLocalProvider(r).FirstID(ctx)
		if repo.Edges.LocalProvider.ID == 0 && err != nil {
			err = createLocalProvider(db, repo.Edges.LocalProvider, newId, ctx)
		} else {
			err = updateLocalProvider(db, repo.Edges.LocalProvider, ctx)
		}
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// update MinioS3Provider
	if repo.Edges.MinioS3Provider != nil {
		if repo.Edges.MinioS3Provider.ID == 0 {
			err = createMinioS3Provider(db, repo.Edges.MinioS3Provider, newId, ctx)
		} else {
			err = updateMinioS3Provider(db, repo.Edges.MinioS3Provider, ctx)
		}
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	err = updatePaths(db, repo.Edges.Paths, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	// Fetch repository after update
	r, err = db.Repository.
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

	r.Valid, _ = validRepository(*r)
	_, err = db.Repository.UpdateOneID(newId).SetValid(r.Valid).Save(ctx)
	if err != nil {
		log.Err(err).Msg("Could not update valid for repository")
	}

	if r.Active && r.Valid {
		addScheduledBackup(r)
	} else {
		removeScheduledBackup(r)
	}

	c.IndentedJSON(http.StatusOK, r)
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

	db := Database.Connect()
	defer db.Close()

	// Query repository to remove backup Schedule
	log.Debug().Msg("Query repository to remove backup Schedule")
	repository := db.Repository.
		Query().
		WithBackupSchedule().
		FirstX(ctx)

	if repository != nil {
		// Remove backup scheduled
		removeScheduledBackup(repository)

		// Delete repository
		log.Debug().Msg("Delete repository")
		err := db.Repository.
			DeleteOneID(newId).
			Exec(ctx)
		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			// c.IndentedJSON(http.StatusNotFound, gin.H{"error": errNotDeleted, "message": err.Error()})
		}
	}

	c.Status(http.StatusNoContent)
}

// LogsRepository godoc
// @Summary logs repository
// @Schemes
// @Description return the logs of the repository
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/logs [get]
func LogsRepository(c *gin.Context) {
	log.Debug().Msg("LogsRepository()")

	ctx := context.Background()
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	db := Database.Connect()
	defer db.Close()

	logsR, err := db.Logs.Query().Where(logs.HasRepositoryWith(repository.ID(newId))).All(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errQuery, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, logsR)
}

// ValidRepository godoc
// @Summary valid repository
// @Schemes
// @Description valid the repository by ID
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/valid [get]
func ValidRepository(c *gin.Context) {
	log.Debug().Msg("ValidRepository()")

	ctx := context.Background()
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
		First(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errQuery, "message": err.Error()})
		return
	}

	if repository != nil {
		valid, errValid := validRepository(*repository)
		if !valid {
			log.Debug().Msg("r not valid")
			_, err := db.Repository.
				UpdateOneID(newId).
				SetValid(valid).
				SetActive(false).
				Save(ctx)
			if err != nil {
				log.Err(err).Msg("Could not update valid and active for repository")
			}
			c.IndentedJSON(http.StatusBadRequest, gin.H{"valid": valid, "error": errValid})
			return
		} else {
			log.Debug().Msg("r valid")
			_, err := db.Repository.
				UpdateOneID(newId).
				SetValid(valid).
				Save(ctx)
			if err != nil {
				log.Err(err).Msg("Could not update valid for repository")
			}
			c.IndentedJSON(http.StatusOK, gin.H{"valid": valid, "error": errValid})
			return
		}

	} else {
		log.Err(errors.New("repository is nil"))
		c.IndentedJSON(http.StatusBadRequest, gin.H{"valid": false, "error": errors.New("repository is nil")})
	}

	c.Status(http.StatusOK)
}

// ValidProvider godoc
// @Summary valid provider of the repository
// @Schemes
// @Description valid the repository by ID
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/provider/valid [get]
func ValidProvider(c *gin.Context) {
	log.Debug().Msg("ValidProvider()")

	ctx := context.Background()
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
		WithAWSS3Provider().
		WithB2Provider().
		WithLocalProvider().
		WithMinioS3Provider().
		Where(repository.ID(newId)).
		First(ctx)

	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errQuery, "message": err.Error()})
		return
	}

	if repository != nil {
		valid, err := validProvider(*repository)
		if err != nil {
			log.Err(err).Msgf("Failed to validate provider")
		} else {
			err = updateValidProvider(db, *repository, valid)
			log.Err(err)
		}
		c.IndentedJSON(http.StatusOK, gin.H{"valid": valid})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"valid": false})
	}

	c.Status(http.StatusOK)
}

func validRepository(repository ent.Repository) (valid bool, err error) {
	log.Debug().Msg("validRepository()")

	valid = (repository.Password != "")
	if !valid { // If password is empty, repository cannot be valid
		log.Debug().Msg("Password empty, repository cannot be valid")
		return false, errors.New("Password empty, repository cannot be valid")
	}
	valid, err = validProvider(repository)
	if err != nil {
		log.Err(err).Msg("Valid repository has failed on provider check")
		return false, err
	}
	if !valid { // If provider is not valid, repository cannot be valid
		return false, nil
	}
	valid, err = validSchedule(repository.Edges.BackupSchedule)
	if err != nil {
		log.Err(err).Msg("Valid repository has failed on schedule check")
		return false, err
	}
	if !valid { // If schedule is not valid, repository cannot be valid
		return false, nil
	}
	return valid, err
}

func validProvider(repository ent.Repository) (valid bool, err error) {
	log.Debug().Msg("validProvider()")

	switch repository.Provider {
	case "localProvider":
		valid, err = validLocalProvider(repository.Edges.LocalProvider)
	case "minioS3Provider":
		valid, err = validMinioS3Provider(repository.Edges.MinioS3Provider)
	case "AWSS3Provider":
		valid, err = validAWSS3Provider(repository.Edges.AWSS3Provider)
	case "B2Provider":
		valid, err = validB2Provider(repository.Edges.B2Provider)
	default:
		valid = false
		err = errors.New("provider does not exist")
	}
	if err != nil {
		log.Err(err).Msg("Error in validation of provider")
	}
	return valid, err
}

func updateValidProvider(db *ent.Client, repository ent.Repository, valid bool) (err error) {
	log.Debug().Msg("updateValidProvider()")
	ctx := context.Background()

	switch repository.Provider {
	case "localProvider":
		if repository.Edges.LocalProvider != nil {
			_, err = db.LocalProvider.UpdateOneID(repository.Edges.LocalProvider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	case "minioS3Provider":
		if repository.Edges.MinioS3Provider != nil {
			_, err = db.MinioS3Provider.UpdateOneID(repository.Edges.MinioS3Provider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	case "AWSS3Provider":
		if repository.Edges.AWSS3Provider != nil {
			_, err = db.AWSS3Provider.UpdateOneID(repository.Edges.AWSS3Provider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	case "B2Provider":
		if repository.Edges.B2Provider != nil {
			_, err = db.B2Provider.UpdateOneID(repository.Edges.B2Provider.ID).SetValid(valid).Save(ctx)
		} else {
			err = fmt.Errorf("%v is nil", repository.Provider)
		}
	default:
		err = errors.New("provider does not exist")
	}
	if err != nil {
		log.Err(err).Msg("Could not update valid for repository")
	}

	return err
}
