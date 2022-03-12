package API

import (
	"context"
	"encoding/json"
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
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"gitlab.com/resticgui/resticgui/internal/commands"
)

const entity = "repository "
const errNotFound = entity + "not found"
const errIDNotValid = "id not valid"
const errNotBind = entity + "bind error"
const errUpdate = entity + "update error"
const errCreate = entity + "creation failed"
const errQuery = entity + "query failed"

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
		// removeScheduledBackup(repository)

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

// StatsRepository godoc
// @Summary Stats repository
// @Schemes
// @Description return the stats command output for the repository
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/stats [get]
func StatsRepository(c *gin.Context) {
	log.Debug().Msg("StatsRepository()")

	ctx := context.Background()
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	db := Database.Connect()
	defer db.Close()

	r, err := db.Repository.
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

	stdout, stderr, errCmd := commands.StatsRepository(*r, false)

	if errCmd == nil {
		db.Logs.
			Create().
			SetOperation("StatsRepository").
			SetStdout(stdout).
			SetStderr(stderr).
			SetError("").
			SetRepositoryID(r.ID).
			Save(ctx)
	} else {
		log.Err(errCmd)
		db.Logs.
			Create().
			SetOperation("StatsRepository").
			SetStdout(stdout).
			SetStderr(stderr).
			SetError(errCmd.Error()).
			SetRepositoryID(r.ID).
			Save(ctx)
	}

	c.IndentedJSON(http.StatusOK, stdout)
}

// CheckRepository godoc
// @Summary Check repository
// @Schemes
// @Description return the check command output for the repository
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/check [get]
func CheckRepository(c *gin.Context) {
	log.Debug().Msg("CheckRepository()")

	ctx := context.Background()
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	db := Database.Connect()
	defer db.Close()

	r, err := db.Repository.
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

	stdout, stderr, errCmd := commands.CheckIntegrity(*r, false, 0)

	if errCmd == nil {
		db.Logs.
			Create().
			SetOperation("CheckIntegrity").
			SetStdout(stdout).
			SetStderr(stderr).
			SetError("").
			SetRepositoryID(r.ID).
			Save(ctx)
	} else {
		log.Err(errCmd)
		db.Logs.
			Create().
			SetOperation("CheckIntegrity").
			SetStdout(stdout).
			SetStderr(stderr).
			SetError(errCmd.Error()).
			SetRepositoryID(r.ID).
			Save(ctx)
	}

	c.IndentedJSON(http.StatusOK, stdout)
}

// GetSnapshotsRepository godoc
// @Summary get snapshots repository
// @Schemes
// @Description return the snapshots command output for the repository
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/snapshots [get]
func GetSnapshotsRepository(c *gin.Context) {
	log.Debug().Msg("GetSnapshotsRepository()")

	ctx := context.Background()
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	db := Database.Connect()
	defer db.Close()

	r, err := db.Repository.
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

	stdout, stderr, errCmd := commands.GetSnapshots(*r)

	if errCmd == nil {
		db.Logs.
			Create().
			SetOperation("GetSnapshots").
			SetStdout(stdout).
			SetStderr(stderr).
			SetError("").
			SetRepositoryID(r.ID).
			Save(ctx)
	} else {
		log.Err(errCmd)
		db.Logs.
			Create().
			SetOperation("GetSnapshots").
			SetStdout(stdout).
			SetStderr(stderr).
			SetError(errCmd.Error()).
			SetRepositoryID(r.ID).
			Save(ctx)
	}

	var snapshots []Models.Snapshot

	err = json.Unmarshal([]byte(stdout), &snapshots)

	if err != nil {
		c.IndentedJSON(http.StatusOK, snapshots)
		return
	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "error converting snapshots to json", "message": err.Error()})
	}

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
	// valid, err = validSchedule(repository.Edges.BackupSchedule)
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
