package api

import (
	"context"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"gitlab.com/resticgui/resticgui/internal/database"
)

// CreatePaths godoc
// @Summary create backup paths
// @Schemes
// @Description create paths of the repository
// @Accept json
// @Param id path int true "Repository ID"
// @Param paths body ent.Path true "Create paths"
// @Produce json
// @Success 201 {object} ent.Path
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/paths [post]
func createPaths(c *gin.Context) {
	log.Debug().Msg("CreatePaths()")

	ctx := context.Background()
	var paths []*ent.Path
	var p []*ent.Path

	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&paths); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	newId, _ := strconv.Atoi(id)

	_, err := database.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
		} else if ent.IsConstraintError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errRConstraint})
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		}
		return
	}

	// create paths
	log.Debug().Msg("create paths")
	p, err = database.CreatePathsDB(paths, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, p)
}

// UpdatePaths godoc
// @Summary update paths
// @Schemes
// @Description update paths of the repository
// @Accept json
// @Param id path int true "Repository ID"
// @Param paths body ent.Path true "Update paths"
// @Produce json
// @Success 200 {object} ent.Path
// @Failure 400 {string} errNotBind
// @Failure 500 {string} errCreate
// @Router /repositories/{id}/paths [patch]
func updatePaths(c *gin.Context) {
	log.Debug().Msg("UpdatePaths()")

	ctx := context.Background()
	var paths []*ent.Path
	var p []*ent.Path

	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	if err := c.BindJSON(&paths); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	newId, _ := strconv.Atoi(id)

	_, err := database.GetRepositoryByIDDB(newId, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errRNotFound})
		} else if ent.IsConstraintError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errRConstraint})
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": errUpdate, "message": err.Error()})
		}
		return
	}

	p, err = database.UpdatePathsDB(paths, newId, ctx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errCreate, "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, p)
}

// GetPath godoc
// @Summary get path info
// @Schemes
// @Description returns files and directories for the provided path
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /paths [post]
func GetPath(c *gin.Context) {
	var root ent.Path

	if err := c.BindJSON(&root); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Error().Err(err).Msg("Bind Error root path")
		return
	}

	log.Info().Msg(root.Name)

	files, err := os.ReadDir(root.Name)
	if err != nil {
		log.Error().Err(err).Msg("File error")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "error in path"})
		return
	}
	paths := []Models.PathStruct{}

	for _, file := range files {
		var path Models.PathStruct
		path.RelPath = root.RelPath
		path.Name = file.Name()
		path.IsDir = file.IsDir()
		path.AbsPath, _ = filepath.Abs(file.Name())
		paths = append(paths, path)
	}

	c.IndentedJSON(http.StatusOK, paths)
}

// GetAllPath godoc
// @Summary get recursive path info
// @Schemes
// @Description returns files and directories recursively starting from the provided path
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /paths/all [post]
func GetAllPath(c *gin.Context) {
	var root ent.Path

	if err := c.BindJSON(&root); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg("Bind Error root path")
		return
	}
	_, err := os.ReadDir(root.Name)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "error in path"})
		return
	}
	paths := []Models.PathStruct{}

	filepath.WalkDir(root.Name, func(pathName string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		var path Models.PathStruct
		path.Name = file.Name()
		path.IsDir = file.IsDir()
		path.AbsPath, _ = filepath.Abs(file.Name())
		paths = append(paths, path)

		return nil
	})

	c.IndentedJSON(http.StatusOK, paths)
}
