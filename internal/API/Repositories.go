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

// QueryRepositories godoc
// @Summary get all repositories
// @Schemes
// @Description returns all respositories
// @Accept json
// @Produce json
// @Success 200 {array} Models.Repository
// @Router /repositories [get]
func QueryRepositories(c *gin.Context) {
	repositories := []Models.Repository{}

	db := Database.Connect()
	db.Preload(clause.Associations).Find(&repositories)

	c.IndentedJSON(http.StatusOK, repositories)
}

// GetRepository godoc
// @Summary get repository
// @Schemes
// @Description returns a respository filtered by his ID
// @Accept json
// @Produce json
// @Param id path int true "Repository ID"
// @Success 200 {object} Models.Repository
// @Failure 404 {string} errIDNotValid
// @Router /repositories/{id} [get]
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

// GetRepository godoc
// @Summary create repository
// @Schemes
// @Description create a repository
// @Accept json
// @Param repository body Models.Repository true "Add account"
// @Produce json
// @Success 201 {object} Models.Repository
// @Failure 400 {string} errNameUsed
// @Failure 500 {string} errCreate
// @Router /repositories [post]
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
	c.IndentedJSON(http.StatusCreated, repository)
}

// GetRepository godoc
// @Summary edit repository
// @Schemes
// @Description edit a repository by ID
// @Accept json
// @Produce json
// @Param id path int true "Repository ID"
// @Param repository body Models.Repository true "Edit account"
// @Success 200 {object} Models.Repository
// @Failure 400 {string} errBindRepository
// @Failure 404 {string} errIDNotValid
// @Router /repositories/{id} [patch]
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
		log.Info().Msg(errNotFound)
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": errNotFound})
		return
	}

	db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&repository).Association("Paths").Replace(&editedRepository.Paths)
	db.Model(&repository.B2Provider).Updates(&editedRepository.B2Provider)
	db.Model(&repository.AWSS3Provider).Updates(&editedRepository.AWSS3Provider)
	db.Model(&repository.LocalProvider).Updates(&editedRepository.LocalProvider)
	db.Model(&repository.MinioS3Provider).Updates(&editedRepository.MinioS3Provider)
	db.Model(&repository.BackupSchedule).Updates(&editedRepository.BackupSchedule)
	db.Model(&repository.BackupPolicy).Updates(&editedRepository.BackupPolicy)
	result = db.Model(repository).Updates(&editedRepository)
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
	c.IndentedJSON(http.StatusOK, updatedRepository)
}

// RemoveRepository godoc
// @Summary delete repository
// @Schemes
// @Description delete a repository by ID
// @Param id path int true "Repository ID"
// @Accept json
// @Produce json
// @Success 204
// @Failure 400 {string} errIDNotValid
// @Failure 404 {string} errNotFound
// @Failure 500 {string} errNotDeleted
// @Router /repositories/{id} [delete]
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
