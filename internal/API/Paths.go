package API

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gitlab.com/resticgui/resticgui/internal/Models"
)

func GetPath(c *gin.Context) {
	//root := c.Param("root") // root path
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

// func AddEntities(c *gin.Context) {
// 	id := c.Param("id") // id of the repository

// 	var repository Models.Repository

// 	db := Database.Connect()

// 	result := db.Preload("Entities").Find(&repository, id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
// 		return
// 	}

// 	// Here we add repository entities
// 	var entities []Models.Entity
// 	err := db.Model(&repository).Association("Entities").Clear()
// 	if err != nil {
// 		log.Info().Err(err).Msg("Clear Error entities")
// 		return
// 	}
// 	err = db.Model(&repository).Association("Entities").Delete(&entities)

// 	if err != nil {
// 		log.Info().Err(err).Msg("Delete Error entities")
// 		return
// 	}

// 	if err := c.BindJSON(&entities); err != nil {
// 		log.Info().Err(err).Msg("Bind Error entities")
// 		return
// 	}

// 	for _, entity := range entities {
// 		Models.Entity.ToString(entity)
// 		//log.Debug().Msgf("entity")
// 	}

// 	checkEntitiesPath(entities)
// 	repository.Entities = entities

// 	if repository.Entities == nil {
// 		log.Debug().Msg("Reset Entities to []")
// 		repository.Entities = []Models.Entity{}
// 		db.Save(&repository)
// 	}
// 	db.Save(&repository)

// 	// Check if the file exist ?

// 	c.IndentedJSON(http.StatusOK, repository.Entities)
// }

// func checkFileExist(path string) bool {
// 	log.Debug().Msgf("path = %v", path)
// 	if _, err := os.Stat(path); err != nil {
// 		if errors.Is(err, os.ErrNotExist) {
// 			// file does not exist
// 			log.Info().Msg("File does not exist")
// 			return false
// 		} else {
// 			// other error
// 			log.Info().Err(err).Msg("File error")
// 			return false
// 		}
// 	}
// 	log.Info().Msg("File exists")
// 	return true
// }
// func checkEntitiesPath(entities []Models.Entity) {
// 	for i, entity := range entities {
// 		entities[i].Exists = checkFileExist(entity.Path)
// 	}
// }

// func GetEntitiesRepository(c *gin.Context) {
// 	id := c.Param("id") // id of the repository

// 	var repository Models.Repository

// 	db := Database.Connect()

// 	result := db.Preload("Entities").Find(&repository, "id = ?", id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
// 		return
// 	}

// 	if repository.Entities == nil {
// 		repository.Entities = []Models.Entity{}
// 	}

// 	c.IndentedJSON(http.StatusOK, repository.Entities)
// }

// func QueryEntities(c *gin.Context) {
// 	var entities []Models.Entity

// 	db := Database.Connect()

// 	db.Find(&entities)

// 	c.IndentedJSON(http.StatusOK, entities)
// }

// func GetEntities(c *gin.Context) {
// 	id := c.Param("id")

// 	var entities Models.Entity

// 	db := Database.Connect()

// 	result := db.Find(&entities, id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "entities not found"})
// 		return
// 	}
// 	c.IndentedJSON(http.StatusOK, entities)
// }

// func CreateEntity(c *gin.Context) {
// 	var entity Models.Entity

// 	if err := c.BindJSON(&entity); err != nil {
// 		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		log.Info().Err(err).Msg("Bind Error Entity")
// 		return
// 	}
// 	db := Database.Connect()

// 	result := db.Create(&entity)
// 	if result.Error != nil {
// 		log.Debug().Err(result.Error)
// 		err := db.Find(&entity).Error

// 		if err == nil {
// 			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "email already in use"})
// 			return
// 		} else {
// 			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "creation failed"})
// 			return
// 		}
// 	}

// 	c.IndentedJSON(http.StatusCreated, result)
// }

// func EditEntity(c *gin.Context) {
// 	id := c.Param("id")

// 	var editedEntity Models.Entity

// 	if err := c.BindJSON(&editedEntity); err != nil {
// 		log.Info().Err(err).Msg("Bind Error Entity")
// 		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var entitys Models.Entity

// 	db := Database.Connect()

// 	result := db.Find(&entitys, id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "entity not found"})
// 		return
// 	}

// 	result = db.Model(&entitys).Updates(&editedEntity)
// 	if result.Error != nil {
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": result.Error.Error(), "message": "Error Updating"})
// 		log.Info().Err(result.Error).Msg("Error updating")
// 		return
// 	}

// 	var updatedEntity Models.Entity

// 	result = db.Find(&updatedEntity, id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "entity not found"})
// 		return
// 	}
// 	c.IndentedJSON(http.StatusOK, updatedEntity)
// }

// func RemoveEntity(c *gin.Context) {
// 	id := c.Param("id")

// 	var entity Models.Entity

// 	db := Database.Connect()

// 	result := db.Find(&entity, id)
// 	if result.RowsAffected == 0 { // Not found entity by id
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "entity not found"})
// 		return
// 	}
// 	result = db.Delete(&entity, id)
// 	if result.Error != nil {
// 		log.Error().Err(result.Error)
// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "entity not deleted"})
// 		return
// 	}
// 	c.Status(http.StatusNoContent)
// }
