package API

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// define router
	router := gin.Default()
	//grp := router.Group("/")
	//{
	//	grp.StaticFS("", http.Dir("dist/front"))
	//}
	router.Use(static.Serve("/", static.LocalFile("../../dist/front/", true)))

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

	// front route
	router.NoRoute(func(c *gin.Context) {
		c.File("../../dist/front/index.html")
	})
	return router
}
