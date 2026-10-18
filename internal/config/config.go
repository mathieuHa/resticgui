package Config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func InitConfig() {
	log.Debug().Msg("InitConfig()")

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		log.Err(err).Msg("Could not read config file")
		panic(log.Panic().Err(err))
	}
}
