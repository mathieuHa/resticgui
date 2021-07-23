package API

import (
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func SetupRouter() *gin.Engine {

	// Activate when going to production
	//gin.SetMode(gin.ReleaseMode)

	// define router
	router := gin.Default()

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
		grpRepositories.GET(":id", GetRepositories)
		grpRepositories.POST("", CreateRepository)
		grpRepositories.PATCH(":id", EditRepository)
		grpRepositories.DELETE(":id", RemoveRepository)
	}

	router.NoRoute(func(c *gin.Context) {
		index := fmt.Sprintf("%s/index.html", viper.GetString("front.path"))
		log.Debug().Msgf("Index= %v", index)
		c.File(index)
	})

	return router
}
