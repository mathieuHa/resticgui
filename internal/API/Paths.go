package API

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
	"gitlab.com/resticgui/resticgui/internal/Models"
)

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

func checkFileExist(path string) bool {
	log.Debug().Msgf("path = %v", path)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file does not exist
			log.Info().Msg("File does not exist")
			return false
		} else {
			// other error
			log.Info().Err(err).Msg("File error")
			return false
		}
	}
	log.Info().Msg("File exists")
	return true
}

func checkAllPath(entities []ent.Path) bool {
	changed := false
	for i, entity := range entities {
		exists := checkFileExist(entity.AbsPath)
		if exists != entities[i].Exists {
			changed = true
		}
		entities[i].Exists = exists
	}
	return changed
}

func updatePaths(db *ent.Client, paths []*ent.Path, repoID int, ctx context.Context) {
	// Remove old path edges
	_, err := db.Repository.
		UpdateOneID(repoID).RemovePaths(db.Repository.Query().QueryPaths().AllX(ctx)...).Save(ctx)
	if err != nil {
		log.Err(err).Msgf("error: %v", errQuery)
		// c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
		// return
	}

	// Update path edges
	if len(paths) > 0 {
		for _, p := range paths {
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
				// c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				// return
			}
			_, err = db.Repository.
				UpdateOneID(repoID).AddPaths(path).Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errQuery)
				// c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				// return
			}
		}
	}

}

func createPaths(db *ent.Client, paths []*ent.Path, repoID int, ctx context.Context) {
	if len(paths) > 0 {
		for _, p := range paths {
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
				// c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				// return
			}
			_, err = db.Repository.
				UpdateOneID(repoID).AddPaths(path).Save(ctx)
			if err != nil {
				log.Err(err).Msgf("error: %v", errQuery)
				// c.IndentedJSON(http.StatusNotFound, gin.H{"error": errCreate, "message": err.Error()})
				// return
			}
		}
	}

}
