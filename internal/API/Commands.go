package API

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/commands"
)

func ExecuteCmd(c *gin.Context) {
	output := ""

	cmd := c.DefaultQuery("cmd", "version")

	log.Info().Msg(cmd)

	switch cmd {
	case "version":
		output = commands.GetResticVersion()
	case "help":
		output = commands.GetResticHelp()
	default:
		output = commands.GetResticVersion()
	}

	c.IndentedJSON(http.StatusOK, output)
}
