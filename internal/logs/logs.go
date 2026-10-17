package Logs

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Initlogs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Set log to debug
	viper.SetDefault("logs.level", int(zerolog.InfoLevel))
	debug := viper.GetBool("logs.debug")
	level := viper.GetInt("logs.level")
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.Level(level))
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Caller().Logger()
	}
	log.Info().Msg("Init Logs Done")
}
