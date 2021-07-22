package Utils

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func TestConfig() {
	viper.SetConfigName("config")         // name of config file (without extension)
	viper.SetConfigType("yaml")           // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("../../testdata") // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		panic(log.Panic().Err(err))
	}
}

// Clean database before test
func CleanDatabaseTest(db *gorm.DB, t assert.TestingT) {
	result := db.Exec("DELETE FROM repositories")
	assert.Nil(t, result.Error)
}
