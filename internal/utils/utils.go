package utils

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func TestConfig() {
	viper.SetConfigName("config")         // name of config file (without extension)
	viper.SetConfigType("yaml")           // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("../../testdata") // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		panic(log.Panic().Err(err))
	}
}

// // Clean database before test
// func CleanDatabaseTest(db *gorm.DB, t assert.TestingT, table string) {
// 	request := fmt.Sprintf("DELETE FROM %s", table)
// 	result := db.Exec(request)
// 	assert.Nil(t, result.Error)
// }

func IsValidInt(s string) bool {
	_, err := strconv.Atoi(s)
	value := (err == nil)
	if value {
		log.Debug().Msg("ID is valid")
	} else {
		log.Info().Msgf("ID is not a valid int")
	}
	return value
}
