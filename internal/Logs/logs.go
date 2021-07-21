package Logs

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Initlogs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Set log to debug
	debug := viper.GetBool("log.debug")
	level := viper.GetInt("log.level")
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.Level(level))
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Caller().Logger()
	}
	log.Info().Msg("Init Logs Done")
}
