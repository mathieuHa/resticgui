package API

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
)

func QueryUsers(c *gin.Context) {
	var users []Models.User

	db := Database.Connect()

	db.Find(&users)

	c.IndentedJSON(http.StatusOK, users)
}

func GetUsers(c *gin.Context) {
	id := c.Param("id")

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user Models.User

	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg("Bind Error User")
		return
	}
	db := Database.Connect()

	result := db.Create(&user)
	if result.Error != nil {
		log.Debug().Err(result.Error)
		err := db.Find(&user).Error

		if err == nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "email already in use"})
			return
		} else {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "creation failed"})
			return
		}
	}

	c.IndentedJSON(http.StatusCreated, user)
}

func EditUser(c *gin.Context) {
	id := c.Param("id")

	var editedUser Models.User

	if err := c.BindJSON(&editedUser); err != nil {
		log.Info().Err(err).Msg("Bind Error User")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	result = db.Model(&users).Updates(&editedUser)
	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": result.Error.Error(), "message": "Error Updating"})
		log.Info().Err(result.Error).Msg("Error updating")
		return
	}

	var updatedUser Models.User

	result = db.Find(&updatedUser, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedUser)
}

func RemoveUser(c *gin.Context) {
	id := c.Param("id")

	var users Models.User

	db := Database.Connect()

	result := db.Delete(&users, id)

	if result.Error != nil {
		log.Error().Err(result.Error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "user not deleted"})
		return
	}

	c.Status(http.StatusNoContent)
}
