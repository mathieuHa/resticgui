package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/models"
	"gitlab.com/resticgui/resticgui/internal/backup"
	"gitlab.com/resticgui/resticgui/internal/database"
	"gitlab.com/resticgui/resticgui/internal/utils"
)

func checkRepositoryByID(id int, ctx context.Context, c *gin.Context) (err error) {
	_, err = database.GetRepositoryByIDDB(id, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
		} else if ent.IsConstraintError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errRConstraint})
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		}
	}
	return err
}

func getRepositoryByID(id int, ctx context.Context, c *gin.Context) (r *ent.Repository, err error) {
	r, err = database.GetRepositoryByIDDB(id, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
		} else if ent.IsConstraintError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errRConstraint})
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		}
	}
	return r, err
}

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

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	r, err := database.GetRepositoryByIDDB(rID, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
			return
		}
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
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	if repo.Password == "" || len(repo.Password) < 4 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": "Password length is too short"})
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

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	var repo ent.Repository

	if err := c.BindJSON(&repo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	// Update repository
	r, err := database.UpdateRepositoryDB(repo, rID, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
			return
		}
		if ent.IsConstraintError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errRConstraint})
			return
		}
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	r.Valid, _ = backup.ValidRepository(*r)

	// Update repository
	r, err = database.UpdateRepositoryDB(*r, rID, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	updateRepositoryRunning(r, rID)

	if r.Active && r.Valid {
		backup.AddScheduledBackup(r, cr)
	} else {
		// backup.RemoveScheduledBackup(r)
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

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	// Query repository to remove backup Schedule
	log.Debug().Msg("Query repository to remove backup Schedule")
	repository, err := database.GetRepositoryByIDDB(rID, ctx)

	if repository != nil && err == nil {
		// Remove backup scheduled
		// todo
		//backup.RemoveScheduledBackup(repository)

		// Delete repository
		log.Debug().Msg("Delete repository")
		err := database.DeleteRepositoryDB(rID, ctx)
		if err != nil {
			log.Err(err).Msgf("error: %v", errQuery)
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
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}
	limit := 10
	page := 0
	limitStr := c.DefaultQuery("limit", "10")
	if utils.IsValidInt(limitStr) {
		limit, _ = strconv.Atoi(limitStr)
	}
	pageStr := c.DefaultQuery("page", "0")
	if utils.IsValidInt(pageStr) {
		page, _ = strconv.Atoi(pageStr)
	}
	_, err = database.GetRepositoryByIDDB(rID, ctx)

	if err != nil && ent.IsNotFound(err) {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
		return
	}
	logsR, err := database.QueryLogsByRepositoryDB(rID, limit, page, ctx)

	if err != nil {
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

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	r, err := database.GetRepositoryByIDDB(rID, ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err, "message": err.Error()})
		return
	}

	r.Valid, err = backup.ValidRepository(*r)
	if !r.Valid {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Repository is not valid", "message": err.Error()})
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

	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	r, err := database.GetRepositoryByIDDB(rID, ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
			return
		}
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	r.Valid, err = backup.ValidRepository(*r)
	if !r.Valid {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRInvalid, "message": err.Error()})
		return
	}

	stdout := backup.CheckRepository(r)

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
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	limit := 0
	page := 0
	limitStr := c.DefaultQuery("limit", "0")
	if utils.IsValidInt(limitStr) {
		limit, _ = strconv.Atoi(limitStr)
		limit = int(math.Abs(float64(limit)))
	}
	pageStr := c.DefaultQuery("page", "0")
	if utils.IsValidInt(pageStr) {
		page, _ = strconv.Atoi(pageStr)
		page = int(math.Abs(float64(page)))
	}

	r, err := database.GetRepositoryByIDDB(rID, ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errUpdate, "message": err.Error()})
		return
	}

	r.Valid, err = backup.ValidRepository(*r)
	if !r.Valid {
		log.Debug().Msg(errRInvalid)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRInvalid, "message": err.Error()})
		return
	}

	stdout := backup.GetSnapshotsRepository(r)

	var snapshots []Models.Snapshot
	var snapshotExtract []Models.Snapshot
	err = json.Unmarshal([]byte(stdout), &snapshots)

	if err == nil {
		if limit != 0 {
			offset := int(math.Min(float64(page*limit), float64(len(snapshots))))
			nextOffset := int(math.Min(float64((page+1)*limit), float64(len(snapshots))))
			for i := offset; i < nextOffset; i++ {
				snapshotExtract = append(snapshotExtract, snapshots[i])
			}
			if snapshotExtract == nil {
				c.IndentedJSON(http.StatusOK, []Models.Snapshot{})
			} else {
				c.IndentedJSON(http.StatusOK, snapshotExtract)
			}
		} else {
			c.IndentedJSON(http.StatusOK, snapshots)
		}
	} else if snapshots == nil {
		c.IndentedJSON(http.StatusOK, []Models.Snapshot{})
	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": errRLogsNotJSON})
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
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}

	r, err := database.GetRepositoryByIDDB(rID, ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
			return
		}
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRUnknown, "message": err.Error()})
		return
	}

	if r != nil {
		r.Valid, err = backup.ValidRepository(*r)
		// Update repository
		_, _ = database.UpdateRepositoryDB(*r, rID, ctx)
		if !r.Valid {
			r.Active = false
			// this
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRInvalid, "message": err.Error()})
		} else {
			c.Status(http.StatusOK)
		}

	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errRUnknown, "message": err.Error()})
	}
}
