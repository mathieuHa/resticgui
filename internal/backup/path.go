package backup

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"gitlab.com/resticgui/resticgui/ent"
)

func checkFileExist(path string) bool {
	log.Debug().Msgf("path = %v", path)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file does not exist
			log.Info().Msg("File does not exist")
			return false
		} else {
			// other error
			log.Info().Err(err).Msg("File error")
			return false
		}
	}
	log.Info().Msg("File exists")
	return true
}

func CheckAllPath(entities []ent.Path) bool {
	changed := false
	for i, entity := range entities {
		exists := checkFileExist(entity.AbsPath)
		if exists != entities[i].Exists {
			changed = true
		}
		entities[i].Exists = exists
	}
	return changed
}
