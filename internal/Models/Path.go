package Models

import (
	"github.com/rs/zerolog/log"
)

type Path struct {
	Base

	Name         string `json:"name" binding:"required,min=1"`
	RepositoryID uint   `json:"repositoryID"`
	Exists       bool   `json:"exists"`
	IsDir        bool   `json:"isDir"`
	AbsPath      string `json:"absPath"`
}

type PathStruct struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"isDir"`
	AbsPath string `json:"absPath"`
}

func (e Path) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Name)
}
