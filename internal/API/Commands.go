package API

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/Commands"
)

func ExecuteCmd(c *gin.Context) {
	output := ""

	cmd := c.DefaultQuery("cmd", "version")

	log.Info().Msg(cmd)

	Commands.GetResticVersion()

	c.IndentedJSON(http.StatusOK, output)
}
