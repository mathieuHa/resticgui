package API

import (
	"fmt"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gitlab.com/resticgui/resticgui/docs"
)

var cr *cron.Cron

const basePath = "/api/v1"

func SetupRouter() *gin.Engine {

	// Activate when going to production
	//gin.SetMode(gin.ReleaseMode)
	docs.SwaggerInfo.BasePath = basePath

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
	grpUsers := router.Group(basePath + "/users")
	{
		grpUsers.GET("", QueryUsers)
		grpUsers.GET(":id", GetUsers)
		grpUsers.POST("", CreateUser)
		grpUsers.PATCH(":id", EditUser)
		grpUsers.DELETE(":id", RemoveUser)
	}

	grpRepositories := router.Group(basePath + "repositories")
	{
		grpRepositories.GET("", QueryRepositories)
		grpRepositories.GET(":id", GetRepository)
		grpRepositories.POST("", CreateRepository)
		grpRepositories.PATCH(":id", EditRepository)
		grpRepositories.DELETE(":id", RemoveRepository)
	}

	grpPaths := router.Group(basePath + "paths")
	{
		grpPaths.POST("", GetPath)
		grpPaths.POST("all", GetPath)
	}

	grpCmds := router.Group(basePath + "cmd")
	{
		grpCmds.GET("", ExecuteCmd)
	}

	grpBackupSchedules := router.Group(basePath + "repositories/:id/schedules")
	{
		grpBackupSchedules.PUT("", UpdateBackupSchedule)
		grpBackupSchedules.GET(":idBackupSchedule", GetBackupSchedule)
		grpBackupSchedules.GET("", QueryBackupSchedule)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.NoRoute(func(c *gin.Context) {
		index := fmt.Sprintf("%s/index.html", viper.GetString("front.path"))
		log.Debug().Msgf("Index= %v", index)
		c.File(index)
	})

	return router
}
