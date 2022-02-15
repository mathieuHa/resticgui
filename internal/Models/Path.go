package Models

import (
	"github.com/rs/zerolog/log"
)

type Path struct {
	Base

	Name         string `json:"name" binding:"required,min=1"`
	RepositoryID uint   `json:"repositoryID"`
	Exists       bool   `json:"exists"`
}

func (e Path) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Name)
}
