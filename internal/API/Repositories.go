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

func QueryRepositories(c *gin.Context) {
	var repositories []Models.Repository

	db := Database.Connect()

	db.Preload(clause.Associations).Find(&repositories)

	c.IndentedJSON(http.StatusOK, repositories)
}

func GetRepository(c *gin.Context) {
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	}

	var repositories Models.Repository

	db := Database.Connect()

	result := db.Preload(clause.Associations).Find(&repositories, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, repositories)
}

func CreateRepository(c *gin.Context) {
	var repository Models.Repository

	if err := c.BindJSON(&repository); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Info().Err(err).Msg("Bind Error Repository")
		return
	}
	db := Database.Connect()

	result := db.Create(&repository) // pass pointer of data to Create

	if result.Error != nil {
		if result.Error != nil {
			log.Debug().Err(result.Error)
			err := db.Find(&repository).Error

			if err == nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "name already in use"})
				return
			} else {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "creation failed"})
				return
			}
		}
	}
	c.IndentedJSON(http.StatusCreated, repository)
}

func EditRepository(c *gin.Context) {
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	}

	var editedRepository Models.Repository

	if err := c.BindJSON(&editedRepository); err != nil {
		log.Info().Err(err).Msg("Bind Error Repository")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var repository Models.Repository

	db := Database.Connect()
	result := db.Preload(clause.Associations).Find(&repository, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&repository).Association("IncludedPath").Replace(&editedRepository.IncludedPath)
	db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&repository).Association("ExcludedPath").Replace(&editedRepository.ExcludedPath)
	db.Model(&repository.BackupSchedule).Updates(&editedRepository.BackupSchedule)
	result = db.Model(&repository).Updates(&editedRepository)
	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": result.Error.Error(), "message": "Error Updating 1"})
		log.Info().Err(result.Error).Msg("Error updating 1")
		return
	}

	var updatedRepository Models.Repository

	result = db.Preload(clause.Associations).Find(&updatedRepository, id)
	if result.RowsAffected == 0 { // Not found entity by id
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedRepository)
}

func RemoveRepository(c *gin.Context) {
	id := c.Param("id")

	if !Utils.IsValidInt(id) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
		return
	}

	var repository Models.Repository

	db := Database.Connect()

	result := db.Find(&repository, id)
	if result.RowsAffected == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	removeScheduledBackup(repository)

	result = db.Select(clause.Associations).Delete(&repository, id)
	if result.Error != nil {
		log.Error().Err(result.Error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "repository not deleted"})
		return
	}
	c.Status(http.StatusNoContent)
}
