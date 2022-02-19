package API

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Models"
)

func GetPath(c *gin.Context) {
	root := c.PostForm("root")
	log.Info().Msgf("%v", root)
	files, err := os.ReadDir(root)
	if err != nil {
		log.Error().Err(err).Msg("File error")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "error in path"})
		return
	}
	var paths []Models.PathStruct

	for _, file := range files {
		var path Models.PathStruct
		path.Name = file.Name()
		path.IsDir = file.IsDir()
		path.AbsPath, _ = filepath.Abs(file.Name())
		paths = append(paths, path)
	}

	c.IndentedJSON(http.StatusOK, paths)
}

func GetPathAll(c *gin.Context) {
	root := c.PostForm("root")
	_, err := os.ReadDir(root)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "error in path"})
		return
	}
	var paths []Models.Path
	filepath.WalkDir(root, func(pathName string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		var path Models.Path
		path.Name = file.Name()
		path.IsDir = file.IsDir()
		path.Exists = true
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
func checkAllPath(entities []Models.Path) {
	for i, entity := range entities {
		entities[i].Exists = checkFileExist(entity.AbsPath)
	}
}
