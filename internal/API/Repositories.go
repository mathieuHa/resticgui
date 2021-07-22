package API

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"net/http"
)

func QueryRepositories(c *gin.Context) {
	var users []Models.Repository

	db := Database.Connect()

	db.Find(&users)

	c.IndentedJSON(http.StatusFound, users)
}

func GetRepositories(c *gin.Context) {
	id := c.Param("id")
	if !Utils.IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var users Models.Repository

	db := Database.Connect()

	result := db.Find(&users, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

func CreateRepository(c *gin.Context) {
	var user Models.Repository

	if err := c.BindJSON(&user); err != nil {
		log.Info().Msg("Bind Error Repository")
		return
	}
	db := Database.Connect()

	result := db.Create(&user) // pass pointer of data to Create
	if result.Error != nil {
		log.Debug().Err(result.Error).Msg("is error?")
		err := db.Find(&user).Error
		if err == nil {
			log.Debug().Err(err).Msg("is error2?")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "email already in use"})
			return
		}
	}
	c.IndentedJSON(http.StatusCreated, user)
}

func EditRepository(c *gin.Context) {
	id := c.Param("id")
	if !Utils.IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var users Models.Repository

	db := Database.Connect()

	result := db.Find(&users, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	// check on invalid data to do
	//errors.Is(result.Error, gorm.ErrInvalidData)

	db.Save(&users)
	c.IndentedJSON(http.StatusOK, users)
}

func RemoveRepository(c *gin.Context) {
	id := c.Param("id")
	if !Utils.IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var users Models.Repository

	db := Database.Connect()

	result := db.Find(&users, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	result = db.Delete(&users, "id = ?", id)
	if result.Error != nil {
		log.Error().Err(result.Error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "user not deleted"})
		return
	}
	c.Status(http.StatusNoContent)
}
