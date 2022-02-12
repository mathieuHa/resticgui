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
		log.Info().Msg("Bind Error User")
		return
	}
	db := Database.Connect()

	result := db.Create(&user) // pass pointer of data to Create
	if result.Error != nil {
		log.Debug().Err(result.Error)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": result.Error})
		return
		// err := db.Find(&user).Error

		// if err == nil {
		// 	log.Debug().Err(err).Msg("is error2?")
		// 	c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "email already in use"})
		// 	return
		// }
	}
	c.IndentedJSON(http.StatusCreated, user)
}

func EditUser(c *gin.Context) {
	id := c.Param("id")

	var editedUser Models.User

	if err := c.BindJSON(&editedUser); err != nil {
		log.Info().Msg("Bind Error User")
		return
	}

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	db.Save(&editedUser)
	c.IndentedJSON(http.StatusOK, editedUser)
}

func RemoveUser(c *gin.Context) {
	id := c.Param("id")

	var users Models.User

	db := Database.Connect()

	result := db.Find(&users, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	result = db.Delete(&users, id)
	if result.Error != nil {
		log.Error().Err(result.Error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "user not deleted"})
		return
	}
	c.Status(http.StatusNoContent)
}
