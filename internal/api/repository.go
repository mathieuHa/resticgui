package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/backup"
	"gitlab.com/resticgui/resticgui/internal/database"
	Models "gitlab.com/resticgui/resticgui/internal/models"
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
	var repoPassword Models.RepositoryPassword

	if c.Request.Body == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": "invalid request"})
		log.Info().Msg(errNotBind)
		return
	}

	if err := c.ShouldBindBodyWith(&repo, binding.JSON); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	if err := c.ShouldBindBodyWith(&repoPassword, binding.JSON); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}
	// Needed because repo does not encode password (sensitive variable)
	repo.Password = repoPassword.Password

	if repo.Password == "" || len(repo.Password) < utils.PasswordMinLength {
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
	updateRepositoryRunning(rID)

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
	limitStr := c.DefaultQuery("limit", fmt.Sprint(limitAPI))
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

	snapshotID := c.DefaultQuery("snapshotID", "")
	background := c.DefaultQuery("background", "false")

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

	if background == "false" || (background != "false" && snapshotID != "") {
		stdout, stderr := backup.GetStatsRepository(r, snapshotID)
		var stat Models.Stat

		err = json.Unmarshal([]byte(stdout), &stat)
		invalidSnapshot := strings.Contains(stderr, "it is not a snapshot id")
		if err == nil && !invalidSnapshot {
			c.IndentedJSON(http.StatusOK, stat)
			return
		}
		if invalidSnapshot {
			c.IndentedJSON(http.StatusOK, gin.H{"error": stderr, "message": "err is not a snapshot id"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": errStatLogsNotJSON})
	} else if snapshotID == "" {
		go func() {
			stdout, stderr := backup.GetStatsRepository(r, snapshotID)
			var stat Models.Stat

			err = json.Unmarshal([]byte(stdout), &stat)
			invalidSnapshot := strings.Contains(stderr, "it is not a snapshot id")
			if err == nil && !invalidSnapshot {
				r.RepositorySize = fmt.Sprint(stat.TotalSize)
				r.RepositoryFileCount = fmt.Sprint(stat.TotalFileCount)
				_, err = database.UpdateRepositoryDB(*r, r.ID, ctx)
				if err != nil {
					log.Err(err).Msg("Err stats in background")
				}
				return
			}
			if invalidSnapshot {
				log.Warn().Msg("Invalid Snapshot ID")
			} else {
				log.Err(err).Msg("Error getting stats")
			}
		}()
		c.IndentedJSON(http.StatusOK, gin.H{"message": "task launched"})
	}

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
	limitStr := c.DefaultQuery("limit", fmt.Sprint(limitAPI))
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
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": errSnapLogsNotJSON})
	}
}

// ForgetSnapshotRepository godoc
// @Summary get snapshots repository
// @Schemes
// @Description return the snapshots command output for the repository
// @Param id path int true "Repository ID"
// @Param snapshotID path string true "Snapshot ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/snapshots/{snapshotID} [delete]
func ForgetSnapshotRepository(c *gin.Context) {
	log.Debug().Msg("ForgetSnapshotRepository()")

	ctx := context.Background()
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}
	snapshotID := c.Param("snapshotID")
	log.Debug().Msgf("SnapshotID: %s", snapshotID)

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

	stdout, stderr := backup.ForgetSnapshotRepository(r, snapshotID)

	c.IndentedJSON(http.StatusOK, gin.H{"stdout": stdout, "stderr": stderr})
}

// RestoreSnapshotRepository godoc
// @Summary restore a snapshot from the repository
// @Schemes
// @Description restore a snapshot from the repository
// @Param id path int true "Repository ID"
// @Param snapshotID path string true "Snapshot ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/restore/{snapshotID} [get]
func RestoreSnapshotRepository(c *gin.Context) {
	log.Debug().Msg("RestoreSnapshotRepository()")

	ctx := context.Background()
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}
	snapshotID := c.Param("snapshotID")
	path := c.DefaultQuery("path", "")
	log.Debug().Msgf("SnapshotID: %s", snapshotID)
	log.Debug().Msgf("path: %s", path)

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

	go func() {
		stdout, stderr := backup.RestoreRepository(r, path, snapshotID)
		log.Debug().Msgf("stdout: %v, stderr: %v", stdout, stderr)
	}()

	c.IndentedJSON(http.StatusOK, gin.H{"message": "task launched"})
}

// ListFilesSnapshotRepository godoc
// @Summary list files present in a snapshot from the repository
// @Schemes
// @Description list files present in a snapshot from the repository
// @Param id path int true "Repository ID"
// @Param snapshotID path string true "Snapshot ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/ls/{snapshotID} [get]
func ListFilesSnapshotRepository(c *gin.Context) {
	log.Debug().Msg("ListFilesSnapshotRepository()")

	ctx := context.Background()
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}
	snapshotID := c.Param("snapshotID")
	log.Debug().Msgf("SnapshotID: %s", snapshotID)

	limit := 0
	page := 0
	limitStr := c.DefaultQuery("limit", fmt.Sprint(limitAPI))
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

	stdout, _ := backup.ListFilesSnapshotRepository(r, snapshotID)
	// log.Debug().Msgf("stdout: %v, stderr: %v", stdout, stderr)

	stdOutArray := strings.Split(strings.ReplaceAll(stdout, "\r\n", "\n"), "\n")
	var obj Models.Files
	var files []Models.Files
	var filesExtract []Models.Files
	for _, s := range stdOutArray {
		log.Debug().Msgf("%s", s)
		err := json.Unmarshal([]byte(s), &obj)
		if err == nil && obj.StructType == "node" {
			files = append(files, obj)
		} else {
			log.Err(err).Msg("Could not convert to Modals.Files")
		}
	}

	if files != nil {
		if limit != 0 {
			offset := int(math.Min(float64(page*limit), float64(len(files))))
			nextOffset := int(math.Min(float64((page+1)*limit), float64(len(files))))
			for i := offset; i < nextOffset; i++ {
				filesExtract = append(filesExtract, files[i])
			}
			if filesExtract == nil {
				c.IndentedJSON(http.StatusOK, []Models.Files{})
			} else {
				c.IndentedJSON(http.StatusOK, filesExtract)
			}
		} else {
			c.IndentedJSON(http.StatusOK, files)
		}
	} else {
		log.Debug().Msg("Files == nil")
		log.Debug().Fields(stdout)
		c.IndentedJSON(http.StatusOK, []Models.Files{})
	}
}

// SearchFilesSnapshotRepository godoc
// @Summary search for filename present in a snapshot from the repository
// @Schemes
// @Description search for filename present in a snapshot from the repository
// @Param id path int true "Repository ID"
// @Param snapshotID path string true "Snapshot ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/search/{snapshotID} [get]
func SearchFilesSnapshotRepository(c *gin.Context) {
	log.Debug().Msg("SearchFilesSnapshotRepository()")

	ctx := context.Background()
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}
	snapshotID := c.Param("snapshotID")
	pattern := c.DefaultQuery("pattern", "")
	log.Debug().Msgf("SnapshotID: %s", snapshotID)
	log.Debug().Msgf("pattern: %s", pattern)

	limit := 0
	page := 0
	limitStr := c.DefaultQuery("limit", fmt.Sprint(limitAPI))
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

	stdout, _ := backup.SearchFilesSnapshotRepository(r, snapshotID, pattern)

	var matches []Models.Matches
	// var files []Models.Files
	var filesExtract []Models.Files
	err = json.Unmarshal([]byte(stdout), &matches)

	if err == nil {
		if limit != 0 && len(matches) > 0 {
			offset := int(math.Min(float64(page*limit), float64(len(matches[0].Files))))
			nextOffset := int(math.Min(float64((page+1)*limit), float64(len(matches[0].Files))))
			for i := offset; i < nextOffset; i++ {
				filesExtract = append(filesExtract, matches[0].Files[i])
			}
			if filesExtract == nil {
				c.IndentedJSON(http.StatusOK, []Models.Files{})
			} else {
				c.IndentedJSON(http.StatusOK, filesExtract)
			}
		} else if len(matches) > 0 {
			c.IndentedJSON(http.StatusOK, matches[0].Files)
		} else {
			c.IndentedJSON(http.StatusOK, []Models.Files{})
		}
	} else if matches == nil {
		c.IndentedJSON(http.StatusOK, []Models.Files{})
	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": errSnapLogsNotJSON})
	}
}

// DownloadFilesSnapshotRepository godoc
// @Summary download files/folder present in a snapshot from the repository
// @Schemes
// @Description download files/folder present in a snapshot from the repository
// @Param id path int true "Repository ID"
// @Param snapshotID path string true "Snapshot ID"
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} error
// @Router /repositories/{id}/download/{snapshotID} [get]
func DownloadFilesSnapshotRepository(c *gin.Context) {
	log.Debug().Msg("DownloadFilesSnapshotRepository()")

	ctx := context.Background()
	rID, err := utils.ReturnValidID(c)
	if err != nil {
		return
	}
	snapshotID := c.Param("snapshotID")
	path := c.DefaultQuery("path", "")
	log.Debug().Msgf("SnapshotID: %s", snapshotID)
	log.Debug().Msgf("path: %s", path)

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
	filename := filepath.Base(path)
	dir := "/tmp/download/" + fmt.Sprint(r.ID) + "/"
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Err(err).Msgf("coud not create dir: %s", dir)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": "err creating directory"})
		return
	}
	// Cannot check if file is dir or not
	// if fileInfo.IsDir() {
	// }
	tmpPath := dir + filename
	_, _ = backup.DownloadFilesSnapshotRepository(r, snapshotID, path, tmpPath)
	buff := make([]byte, 512)
	file, err := os.Open(tmpPath)

	if err != nil {
		log.Err(err).Msgf("coud not read file written: %s", tmpPath)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": "err coud not read file written"})
		return
	}

	defer file.Close()
	_, err = file.Read(buff)
	if err != nil {
		log.Err(err).Msgf("coud not read buff")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"err": "coud not read buff"})
		return
	}
	filetype := http.DetectContentType(buff)
	log.Debug().Msgf("filetype: %s", filetype)

	if filetype == "application/zip" {
		filename = filename + ".zip"
	}

	c.FileAttachment(tmpPath, filename)
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
