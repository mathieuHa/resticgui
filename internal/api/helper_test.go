package api

import (
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.com/resticgui/resticgui/internal/Config"
	"gitlab.com/resticgui/resticgui/internal/database"
)

func setupRouterTest(t *testing.T) *gin.Engine {
	t.Helper()

	Config.InitTestConfig()
	database.Migrate()
	database.CleanDatabaseTest(t)
	router := SetupRouter()
	gin.SetMode(gin.ReleaseMode)
	return router
}
