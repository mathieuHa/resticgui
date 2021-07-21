package Logs

import (
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init_logs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Set log to debug
	debug := flag.Bool("debug", true, "sets log level to debug")
	flag.Parse()
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Caller().Logger()
	}
	log.Info().Msg("Init Logs Done")
}
