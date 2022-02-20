package API

import (
	"fmt"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var cr *cron.Cron

func SetupRouter() *gin.Engine {

	// Activate when going to production
	//gin.SetMode(gin.ReleaseMode)

	// define router
	router := gin.Default()

	go func() {
		cr = cron.New()
		cr.Start()
		for _, value := range cr.Entries() {
			log.Info().Msgf("%v", value.ID)
		}
		CreateAllScheduledBackup()
	}()
	router.Use(static.Serve("/", static.LocalFile(viper.GetString("front.path"), true)))
	// API routes
	grpUsers := router.Group("/api/v1/users")
	{
		grpUsers.GET("", QueryUsers)
		grpUsers.GET(":id", GetUsers)
		grpUsers.POST("", CreateUser)
		grpUsers.PATCH(":id", EditUser)
		grpUsers.DELETE(":id", RemoveUser)
	}

	grpRepositories := router.Group("/api/v1/repositories")
	{
		grpRepositories.GET("", QueryRepositories)
		grpRepositories.GET(":id", GetRepository)
		grpRepositories.POST("", CreateRepository)
		grpRepositories.PATCH(":id", EditRepository)
		grpRepositories.DELETE(":id", RemoveRepository)
	}

	grpPaths := router.Group("/api/v1/paths")
	{
		grpPaths.POST("", GetPath)
		grpPaths.POST("all", GetPath)
	}

	grpBackupSchedules := router.Group("/api/v1/repositories/:id/schedules")
	{
		grpBackupSchedules.PUT("", UpdateBackupSchedule)
		grpBackupSchedules.GET(":idBackupSchedule", GetBackupSchedule)
		grpBackupSchedules.GET("", QueryBackupSchedule)
	}

	router.NoRoute(func(c *gin.Context) {
		index := fmt.Sprintf("%s/index.html", viper.GetString("front.path"))
		log.Debug().Msgf("Index= %v", index)
		c.File(index)
	})

	return router
}
