package API

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gorm.io/gorm"
	"net/http"
)

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func QueryUsers(c *gin.Context) {
	var users []Models.User

	db := Database.Connect()

	db.Find(&users)

	c.IndentedJSON(http.StatusFound, users)
}

func GetUsers(c *gin.Context) {
	id := c.Param("id")
	if !IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user Models.User

	if err := c.BindJSON(&user); err != nil {
		log.Info().Msg("Bind Error User")
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

func EditUser(c *gin.Context) {
	id := c.Param("id")
	if !IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	errors.Is(result.Error, gorm.ErrInvalidData)

	db.Save(&users)
	c.IndentedJSON(http.StatusOK, users)
}

func RemoveUser(c *gin.Context) {
	id := c.Param("id")
	if !IsValidUUID(id) {
		log.Info().Msgf("UUID is not valid")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	} else {
		log.Debug().Msg("UUID is valid")
	}

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, "id = ?", id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	db.Delete(&users, "id = ?", id)
	c.Status(http.StatusNoContent)
}
