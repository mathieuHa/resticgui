package backup

import (
	"errors"
	"os"
	"syscall"

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

func validRestorePath(path string) (valid bool, err error) {
	log.Debug().Msg("validRestorePath()")

	log.Info().Msgf("restorePath: %v", path)
	fileinfo, err := os.Stat(path)
	userWrite := true
	// snapshot will be restored in path
	if os.IsNotExist(err) {
		log.Err(err).Msgf("Restore Path %v does not exist", path)
		// can we try to write and create this directory ?
		return false, err
	}
	if err != nil {
		log.Err(err).Msgf("Restore Path %v unknown error", path)
		return false, err
	}
	if !fileinfo.IsDir() {
		log.Err(err).Msgf("Restore Path %v is not a directory", path)
		return false, err
	}
	log.Debug().Msgf("perms : %d %d", fileinfo.Mode().Perm(), 1<<(uint(7)))

	if fileinfo.Mode().Perm()&(0005) != 0 {
		log.Info().Msgf("Restore Path %v is writable by everyone", path)
		return true, nil
	}
	// only see the user that owns the file and not the current
	if fileinfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		log.Info().Msgf("Restore Path %v is not writable by user", path)
		userWrite = false
	}
	// only work on linux ?
	var stat syscall.Stat_t
	if err = syscall.Stat(path, &stat); err != nil {
		log.Err(err).Msgf("Restore Path %v unable to get stat", path)
		return false, err
	}

	if uint32(os.Geteuid()) != stat.Uid {
		log.Info().Msgf("Restore Path %v user doesn't have permission to write to this directory", path)
	} else if userWrite {
		return true, nil
	}

	return false, nil
}

func validPaths(paths []*ent.Path) (valid bool, err error) {
	log.Debug().Msg("validPaths()")
	// Check path len > 0
	if len(paths) == 0 {
		log.Info().Msg("Path list is empty, nothing to backup")
		return false, errors.New("Path list is empty, nothing to backup")
	}
	return true, nil
}
