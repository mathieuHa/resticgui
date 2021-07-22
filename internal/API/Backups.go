package API

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func QueryBackups(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "backups"})
}

func GetBackup(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "backups"})
}

func CreateBackup(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "backups"})
}

func EditBackup(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "backups"})
}

func RemoveBackup(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "backups"})
}
