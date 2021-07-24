package API

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"net/http"
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

	var repositories Models.Repository

	db := Database.Connect()

	result := db.Find(&repositories, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	// Here we add repository entities
	var entities []Models.Entity

	if err := c.BindJSON(&entities); err != nil {
		log.Info().Msg("Bind Error entities")
		return
	}

	err := db.Model(&repositories).Association("Entities").Append(entities)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error adding entities"})
		return
	}

	// Check if the file exist ?

	c.IndentedJSON(http.StatusOK, entities)
}
