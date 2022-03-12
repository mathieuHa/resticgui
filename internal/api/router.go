package api

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
	"gitlab.com/resticgui/resticgui/internal/backup"
)

var cr *cron.Cron

const APIBasePath = "/api/v1/"
const APITitle = "ResticGUI API"
const APIVersion = "1.0"
const APIdescription = "ResticGUI API allows to create and manipulates repositories and backups which translate in restic commands under the hood"
const APIHost = "localhost:3001"

func SetupRouter() *gin.Engine {

	// Activate when going to production
	//gin.SetMode(gin.ReleaseMode)
	docs.SwaggerInfo.BasePath = APIBasePath
	docs.SwaggerInfo.Title = APITitle
	docs.SwaggerInfo.Version = APIVersion
	docs.SwaggerInfo.Description = APIdescription
	docs.SwaggerInfo.Host = APIHost
	// define router
	router := gin.Default()

	go func() {
		cr = cron.New()
		cr.Start()
		for _, value := range cr.Entries() {
			log.Info().Msgf("%v", value.ID)
		}
		backup.CreateAllScheduledBackup(cr)
	}()
	router.Use(static.Serve("/", static.LocalFile(viper.GetString("front.path"), true)))

	grpRepositories := router.Group(APIBasePath + "repositories")
	{
		grpRepositories.GET("", QueryRepositories)
		grpRepositories.GET(":id", GetRepository)
		grpRepositories.POST("", CreateRepository)
		grpRepositories.POST(":id/backupschedule", CreateBackupSchedule)
		grpRepositories.PATCH(":id/backupschedule", UpdateBackupSchedule)
		grpRepositories.POST(":id/backuppolicy", CreateBackupPolicy)
		grpRepositories.PATCH(":id/backuppolicy", UpdateBackupPolicy)
		grpRepositories.POST(":id/b2provider", CreateB2Provider)
		grpRepositories.PATCH(":id/b2provider", UpdateB2Provider)
		grpRepositories.PATCH(":id", EditRepository)
		grpRepositories.DELETE(":id", RemoveRepository)
		grpRepositories.GET(":id/logs", LogsRepository)
		grpRepositories.GET(":id/stats", StatsRepository)
		grpRepositories.GET(":id/check", CheckRepository)
		grpRepositories.GET(":id/snapshots", GetSnapshotsRepository)
		grpRepositories.GET(":id/valid", ValidRepository)
	}

	// grpPaths := router.Group(APIBasePath + "paths")
	// {
	// 	grpPaths.POST("", GetPath)
	// 	grpPaths.POST("all", GetAllPath)
	// }

	// grpCmds := router.Group(APIBasePath + "cmd")
	// {
	// 	grpCmds.GET("", ExecuteCmd)
	// }

	// grpBackupSchedules := router.Group(APIBasePath + "repositories/:id/schedules")
	// {
	// 	grpBackupSchedules.PUT("", UpdateBackupSchedule)
	// 	grpBackupSchedules.GET(":idBackupSchedule", GetBackupSchedule)
	// 	grpBackupSchedules.GET("", QueryBackupSchedule)
	// }

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.NoRoute(func(c *gin.Context) {
		index := fmt.Sprintf("%s/index.html", viper.GetString("front.path"))
		log.Debug().Msgf("Index= %v", index)
		c.File(index)
	})

	return router
}
