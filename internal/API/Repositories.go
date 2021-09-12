package API

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
)

func QueryRepositories(c *gin.Context) {
	var repositories []Models.Repository

	db := Database.Connect()

	db.Find(&repositories)

	c.IndentedJSON(http.StatusOK, repositories)
}

func GetRepositories(c *gin.Context) {
	id := c.Param("id")
	// if !Utils.IsValidUUID(id) {
	// 	log.Info().Msgf("UUID is not valid")
	// 	c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
	// 	return
	// } else {
	// 	log.Debug().Msg("UUID is valid")
	// }

	var repositories Models.Repository

	db := Database.Connect()

	result := db.Preload("Entities").Find(&repositories, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, repositories)
}

func CreateRepository(c *gin.Context) {
	var repository Models.Repository

	if err := c.BindJSON(&repository); err != nil {
		log.Info().Msg("Bind Error Repository")
		return
	}
	db := Database.Connect()

	result := db.Create(&repository) // pass pointer of data to Create

	if result.Error != nil {
		log.Debug().Err(result.Error).Msg("is error?")
		err := db.Find(&repository).Error
		if err == nil {
			log.Debug().Err(err).Msg("is error2?")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "name already in use"})
			return
		}
	}
	// repository.BackupSchedule = Models.BackupSchedule{}
	result = db.Save(&repository)

	c.IndentedJSON(http.StatusCreated, repository)
}

func EditRepository(c *gin.Context) {
	id := c.Param("id")

	var repository Models.Repository

	db := Database.Connect()

	// result := db.Preload("Entities").Find(&repository, id)
	result := db.Find(&repository, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	log.Error().Msgf("DEBUG BEFORE Error Repository %s", repository.Entities)

	if err := c.BindJSON(&repository); err != nil {
		log.Info().Msg("Bind Error Repository")
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not binded"})
		return
	}

	log.Error().Msgf("DEBUG AFTER Error Repository %s", repository.Entities)

	db.Save(repository)

	var NewRepository Models.Repository

	result2 := db.Preload("Entities").Find(&NewRepository, id)
	if result2.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, NewRepository)
}

func RemoveRepository(c *gin.Context) {
	id := c.Param("id")

	var repository Models.Repository

	db := Database.Connect()

	result := db.Find(&repository, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	result = db.Delete(&repository, "id = ?", id)
	if result.Error != nil {
		log.Error().Err(result.Error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "repository not deleted"})
		return
	}
	c.Status(http.StatusNoContent)
}
