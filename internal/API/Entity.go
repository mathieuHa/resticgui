package API

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"net/http"
	"os"
)

// AddEntities This could be redundant with the update repository ?
func AddEntities(c *gin.Context) {
	id := c.Param("id") // id of the repository
	if !Utils.IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var repository Models.Repository

	db := Database.Connect()

	result := db.Preload("Entities").Find(&repository, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	// Here we add repository entities
	var entities []Models.Entity
	if err := c.BindJSON(&entities); err != nil {
		log.Info().Err(err).Msg("Bind Error entities")
		return
	}


	//err := db.Model(&repository).Association("Entities").Replace(entities)
	//if err != nil {
	//	c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error adding entities"})
	//	return
	//}
	checkEntitiesPath(entities)
	repository.Entities = entities
	if repository.Entities == nil {
	    log.Debug().Msg("Reset Entities to []")
        repository.Entities = []Models.Entity{}
    }
	db.Save(&repository)

	// Check if the file exist ?

	c.IndentedJSON(http.StatusOK, repository.Entities)
}

func checkFileExist(path string) bool {
	log.Debug().Msgf("path = %v",path)
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
func checkEntitiesPath(entities []Models.Entity) {
	for i, entity := range entities {
		entities[i].Exists = checkFileExist(entity.Path)
	}
}

func GetEntities(c *gin.Context) {
	id := c.Param("id") // id of the repository
	if !Utils.IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var repository Models.Repository

	db := Database.Connect()

	result := db.Preload("Entities").Find(&repository, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	if repository.Entities == nil {
		repository.Entities = []Models.Entity{}
	}

	c.IndentedJSON(http.StatusOK, repository.Entities)
}
