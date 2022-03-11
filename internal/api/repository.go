package api

import (
	"context"
	"encoding/json"
	"errors"
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
	"gitlab.com/resticgui/resticgui/internal/backup"
	"gitlab.com/resticgui/resticgui/internal/commands"
	"gitlab.com/resticgui/resticgui/internal/database"
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

	repositories, err := database.GetRepositoriesDB(ctx)
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

	r, err := database.GetRepositoryByIDDB(newId, ctx)
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

	r, err := database.CreateRepositoryDB(repo, ctx)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	r.Valid, err = backup.ValidRepository(*r)
	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
	}

	// if r.Active && r.Valid {
	// 	bs.AddScheduledBackup(r)
	// }

	c.IndentedJSON(http.StatusCreated, r)
}

// EditRepository godoc
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
	r, err := database.UpdateRepositoryDB(repo, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	r.Valid, _ = backup.ValidRepository(*r)

	// Update repository
	r, err = database.UpdateRepositoryDB(*r, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	// if r.Active && r.Valid {
	// 	backup.AddScheduledBackup(r, cr)
	// } else {
	// 	backup.RemoveScheduledBackup(r)
	// }

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
		// backup.RemoveScheduledBackup(repository)

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

	r, err := database.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stdout := backup.GetStatsRepository(r)
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
		valid, errValid := backup.ValidRepository(*repository)
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
