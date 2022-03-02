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
		_, err = db.BackupSchedule.
			Create().
			SetChoice(repo.Edges.BackupSchedule.Choice).
			SetCronSyntax(repo.Edges.BackupSchedule.CronSyntax).
			SetDay(repo.Edges.BackupSchedule.Day).
			SetHour(repo.Edges.BackupSchedule.Hour).
			SetEveryDay(repo.Edges.BackupSchedule.EveryDay).
			SetEveryWeek(repo.Edges.BackupSchedule.EveryWeek).
			SetEveryMonth(repo.Edges.BackupSchedule.EveryMonth).
			SetEveryMonday(repo.Edges.BackupSchedule.EveryMonday).
			SetEveryTuesday(repo.Edges.BackupSchedule.EveryTuesday).
			SetEveryWednesday(repo.Edges.BackupSchedule.EveryWednesday).
			SetEveryThursday(repo.Edges.BackupSchedule.EveryThursday).
			SetEveryFriday(repo.Edges.BackupSchedule.EveryFriday).
			SetEverySaturday(repo.Edges.BackupSchedule.EverySaturday).
			SetEverySunday(repo.Edges.BackupSchedule.EverySunday).
			SetRepository(newRepository).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create backupPolicy
	if repo.Edges.BackupPolicy != nil {
		_, err = db.BackupPolicy.
			Create().
			SetKeepAll(repo.Edges.BackupPolicy.KeepAll).
			SetName(repo.Edges.BackupPolicy.Name).
			SetNbLastBackup(repo.Edges.BackupPolicy.NbLastBackup).
			SetOneByDay(repo.Edges.BackupPolicy.OneByDay).
			SetOneByHour(repo.Edges.BackupPolicy.OneByHour).
			SetOneByMonth(repo.Edges.BackupPolicy.OneByMonth).
			SetRepository(newRepository).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create AWSS3Provider
	if repo.Edges.AWSS3Provider != nil {
		_, err = db.AWSS3Provider.
			Create().
			SetAccessKeyID(repo.Edges.AWSS3Provider.AccessKeyID).
			SetSecretAccessKey(repo.Edges.AWSS3Provider.SecretAccessKey).
			SetBucketName(repo.Edges.AWSS3Provider.BucketName).
			SetDefaultRegion(repo.Edges.AWSS3Provider.DefaultRegion).
			SetRepository(newRepository).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create B2Provider
	if repo.Edges.B2Provider != nil {
		_, err = db.B2Provider.
			Create().
			SetAccountID(repo.Edges.B2Provider.AccountID).
			SetAccountKey(repo.Edges.B2Provider.AccountKey).
			SetBucketName(repo.Edges.B2Provider.BucketName).
			SetRepository(newRepository).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create LocalProvider
	if repo.Edges.LocalProvider != nil {
		_, err = db.LocalProvider.
			Create().
			SetRepoPath(repo.Edges.LocalProvider.RepoPath).
			SetRepository(newRepository).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// create MinioS3Provider
	if repo.Edges.MinioS3Provider != nil {
		_, err = db.MinioS3Provider.
			Create().
			SetAccessKeyID(repo.Edges.MinioS3Provider.AccessKeyID).
			SetSecretAccessKey(repo.Edges.MinioS3Provider.SecretAccessKey).
			SetBucketName(repo.Edges.MinioS3Provider.BucketName).
			SetIP(repo.Edges.MinioS3Provider.IP).
			SetPort(repo.Edges.MinioS3Provider.Port).
			SetScheme(repo.Edges.MinioS3Provider.Scheme).
			SetRepository(newRepository).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
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
		_, err = db.BackupSchedule.
			UpdateOneID(repo.Edges.BackupSchedule.ID).
			SetChoice(repo.Edges.BackupSchedule.Choice).
			SetCronSyntax(repo.Edges.BackupSchedule.CronSyntax).
			SetDay(repo.Edges.BackupSchedule.Day).
			SetHour(repo.Edges.BackupSchedule.Hour).
			SetEveryDay(repo.Edges.BackupSchedule.EveryDay).
			SetEveryWeek(repo.Edges.BackupSchedule.EveryWeek).
			SetEveryMonth(repo.Edges.BackupSchedule.EveryMonth).
			SetEveryMonday(repo.Edges.BackupSchedule.EveryMonday).
			SetEveryTuesday(repo.Edges.BackupSchedule.EveryTuesday).
			SetEveryWednesday(repo.Edges.BackupSchedule.EveryWednesday).
			SetEveryThursday(repo.Edges.BackupSchedule.EveryThursday).
			SetEveryFriday(repo.Edges.BackupSchedule.EveryFriday).
			SetEverySaturday(repo.Edges.BackupSchedule.EverySaturday).
			SetEverySunday(repo.Edges.BackupSchedule.EverySunday).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errUpdate, "message": err.Error()})
			return
		}
	}

	// Update backupPolicy
	if repo.Edges.BackupPolicy != nil {
		_, err = db.BackupPolicy.
			UpdateOneID(repo.Edges.BackupPolicy.ID).
			SetKeepAll(repo.Edges.BackupPolicy.KeepAll).
			SetName(repo.Edges.BackupPolicy.Name).
			SetNbLastBackup(repo.Edges.BackupPolicy.NbLastBackup).
			SetOneByDay(repo.Edges.BackupPolicy.OneByDay).
			SetOneByHour(repo.Edges.BackupPolicy.OneByHour).
			SetOneByMonth(repo.Edges.BackupPolicy.OneByMonth).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// Update AWSS3Provider
	if repo.Edges.AWSS3Provider != nil {
		_, err = db.AWSS3Provider.
			UpdateOneID(repo.Edges.AWSS3Provider.ID).
			SetAccessKeyID(repo.Edges.AWSS3Provider.AccessKeyID).
			SetSecretAccessKey(repo.Edges.AWSS3Provider.SecretAccessKey).
			SetBucketName(repo.Edges.AWSS3Provider.BucketName).
			SetDefaultRegion(repo.Edges.AWSS3Provider.DefaultRegion).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// Update B2Provider
	if repo.Edges.B2Provider != nil {
		_, err = db.B2Provider.
			UpdateOneID(repo.Edges.B2Provider.ID).
			SetAccountID(repo.Edges.B2Provider.AccountID).
			SetAccountKey(repo.Edges.B2Provider.AccountKey).
			SetBucketName(repo.Edges.B2Provider.BucketName).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// update LocalProvider
	if repo.Edges.LocalProvider != nil {
		_, err = db.LocalProvider.
			UpdateOneID(repo.Edges.LocalProvider.ID).
			SetRepoPath(repo.Edges.LocalProvider.RepoPath).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
		}
	}

	// update MinioS3Provider
	if repo.Edges.MinioS3Provider != nil {
		_, err = db.MinioS3Provider.
			UpdateOneID(repo.Edges.MinioS3Provider.ID).
			SetAccessKeyID(repo.Edges.MinioS3Provider.AccessKeyID).
			SetSecretAccessKey(repo.Edges.MinioS3Provider.SecretAccessKey).
			SetBucketName(repo.Edges.MinioS3Provider.BucketName).
			SetIP(repo.Edges.MinioS3Provider.IP).
			SetPort(repo.Edges.MinioS3Provider.Port).
			SetScheme(repo.Edges.MinioS3Provider.Scheme).
			Save(ctx)

		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate, "message": err.Error()})
			return
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
