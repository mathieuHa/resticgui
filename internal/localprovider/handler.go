package localprovider

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"gitlab.com/resticgui/resticgui/internal/repository"
)

// CreateLocalProvider godoc
// @Summary create backup local provider
// @Schemes
// @Description create a localprovider
// @Accept json
// @Param id path int true "Repository ID"
// @Param backupschedule body ent.BackupSchedule true "Create localprovider"
// @Produce json
// @Success 201 {object} ent.LocalProvider
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/localprovider [post]
func CreateLocalProvider(c *gin.Context) {
	log.Debug().Msg("CreateLocalProvider()")

	ctx := context.Background()
	var p ent.LocalProvider

	if err := c.BindJSON(&p); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	newId, _ := strconv.Atoi(id)

	_, err := repository.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// create localprovider
	log.Debug().Msg("create localprovider")
	backupPolicy, err := CreateLocalProviderDB(p, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, backupPolicy)
}

// UpdateLocalProvider godoc
// @Summary update local provider
// @Schemes
// @Description update a localprovider
// @Accept json
// @Param id path int true "Repository ID"
// @Param localprovider body ent.LocalProvider true "Update localprovider"
// @Produce json
// @Success 200 {object} ent.LocalProvider
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/localprovider [patch]
func UpdateLocalProvider(c *gin.Context) {
	log.Debug().Msg("UpdateLocalProvider()")

	ctx := context.Background()
	var p ent.LocalProvider
	var lp *ent.LocalProvider

	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&p); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	newId, _ := strconv.Atoi(id)

	_, err := repository.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if p.ID == 0 && err != nil {
		lp, err = CreateLocalProviderDB(p, newId, ctx)
	} else {
		lp, err = UpdateLocalProviderDB(p, ctx)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, lp)
}

func ValidLocalProvider(provider *ent.LocalProvider) (valid bool, err error) {
	log.Debug().Msg("validLocalProvider()")

	if provider == nil {
		valid = false
		err = errors.New("localProvider is nil")
		return valid, err
	}

	log.Info().Msgf("repoPath: %v", provider.RepoPath)
	fileinfo, err := os.Stat(provider.RepoPath)
	if os.IsNotExist(err) {
		log.Err(err).Msgf("Repository Path %v does not exist", provider.RepoPath)
		return false, err
	}
	if !fileinfo.IsDir() {
		log.Err(err).Msgf("Repository Path %v is not a directory", provider.RepoPath)
		return false, err
	}
	if fileinfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		log.Err(err).Msgf("Repository Path %v is not writable by user", provider.RepoPath)
		return false, err
	}

	return true, nil
}
