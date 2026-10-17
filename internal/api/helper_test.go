package api

import (
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mathieuHa/resticgui/internal/config"
	"github.com/mathieuHa/resticgui/internal/database"
)

var (
	routerOnce sync.Once
	testRouter *gin.Engine
)

// setupRouterTest builds the router once per test binary: SetupRouter starts
// the scheduler in a goroutine that reads the config, so rebuilding it per test
// raced with the next test reloading the config.
func setupRouterTest(t *testing.T) *gin.Engine {
	t.Helper()

	routerOnce.Do(func() {
		Config.InitTestConfig()
		database.Migrate()
		gin.SetMode(gin.ReleaseMode)
		testRouter = SetupRouter()
	})
	return testRouter
}

func cleanDataBaseTest(t *testing.T) {
	database.CleanDatabaseTest(t)
}
