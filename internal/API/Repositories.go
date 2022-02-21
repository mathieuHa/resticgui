package API

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Database"
	"gitlab.com/resticgui/resticgui/internal/Models"
	"gitlab.com/resticgui/resticgui/internal/Utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const entity = "repository "
const errNotFound = entity + "not found"
const errIDNotValid = "id not valid"
const errNotDeleted = entity + "not deleted"
const errNotBind = entity + "bind error"
const errUpdate = entity + "update error"
const errNameUsed = entity + "name already in use"
const errCreate = entity + "creation failed"

func QueryRepositories(c *gin.Context) {
	repositories := []Models.Repository{}

	db := Database.Connect()
	db.Preload(clause.Associations).Find(&repositories)

	c.IndentedJSON(http.StatusOK, repositories)
}

func GetRepository(c *gin.Context) {
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	var repositories Models.Repository

	db := Database.Connect()
	result := db.Preload(clause.Associations).Find(&repositories, id)

	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errNotFound})
		return
	}
	// check all path and update entity in memory if path changed
	checkAllPath(repositories.Paths)

	c.IndentedJSON(http.StatusOK, repositories)
}

func CreateRepository(c *gin.Context) {
	var repository Models.Repository

	if err := c.BindJSON(&repository); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg(errNotBind)
		return
	}

	db := Database.Connect()
	result := db.Create(&repository) // pass pointer of data to Create

	if result.Error != nil {
		if result.Error != nil {
			log.Debug().Err(result.Error)
			err := db.Find(&repository).Error

			if err == nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errNameUsed})
				return
			} else {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errCreate})
				return
			}
		}
	}
	// commands.InitRepository(repository)
	c.IndentedJSON(http.StatusCreated, repository)
}

func EditRepository(c *gin.Context) {
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	var editedRepository Models.Repository

	if err := c.BindJSON(&editedRepository); err != nil {
		log.Info().Err(err).Msg(errNotBind)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var repository Models.Repository

	db := Database.Connect()
	result := db.Preload(clause.Associations).Find(&repository, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errNotFound})
		return
	}

	db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&repository).Association("Paths").Replace(&editedRepository.Paths)
	db.Model(&repository.BackupSchedule).Updates(&editedRepository.BackupSchedule)
	result = db.Model(&repository).Updates(&editedRepository)
	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": result.Error.Error(), "message": errUpdate})
		log.Info().Err(result.Error).Msg(errUpdate)
		return
	}

	var updatedRepository Models.Repository

	result = db.Preload(clause.Associations).Find(&updatedRepository, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errNotFound})
		return
	}
	// commands.InitRepository(repository)
	c.IndentedJSON(http.StatusOK, updatedRepository)
}

func RemoveRepository(c *gin.Context) {
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": errIDNotValid})
		return
	}

	var repository Models.Repository

	db := Database.Connect()
	result := db.Find(&repository, id)
	if result.RowsAffected == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errNotFound})
		return
	}

	removeScheduledBackup(repository)

	result = db.Select(clause.Associations).Delete(&repository, id)
	if result.Error != nil {
		log.Error().Err(result.Error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": errNotDeleted})
		return
	}
	c.Status(http.StatusNoContent)
}
