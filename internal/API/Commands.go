package API

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/internal/commands"
)

// ExecuteCmd godoc
// @Summary execute restic cmd
// @Schemes
// @Description execute a restic cmd passed by queryParam
// @Param cmd query string false "cmd" Enums(version, help)
// @Accept json
// @Produce json
// @Success 200 {string} outputCmd
// @Router /cmd [get]
func ExecuteCmd(c *gin.Context) {
	output := ""

	cmd := c.DefaultQuery("cmd", "version")

	log.Info().Msg(cmd)

	switch cmd {
	case "version":
		output = commands.GetResticHelp()
	case "help":
		output = commands.GetResticHelp()
	default:
		output = commands.GetResticHelp()
	}

	c.IndentedJSON(http.StatusOK, output)
}
