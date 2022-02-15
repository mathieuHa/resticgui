package Models

import (
	"time"

	"github.com/rs/zerolog/log"
)

type User struct {
	Base

	Firstname          string     `json:"firstname" binding:"required,min=3"`
	Lastname           string     `json:"lastname" binding:"required,min=3"`
	Email              string     `json:"email" gorm:"unique" binding:"required,email"`
	Password           string     `json:"password"`
	PasswordLastSet    *time.Time `json:"passwordLastSet"`
	PasswordFailCount  uint8      `json:"passwordFailCount"`
	ResetPassword      bool       `json:"resetPassword"`
	LastConnectionDate time.Time  `json:"lastConnectionDate"`
	Mfa                bool       `json:"mfa"`
	MfaToken           string     `json:"mfaToken"`
	Role               Role
	Active             bool `json:"active"`
}

func (e User) ToString() {
	log.Debug().Msgf("ID = %v", e.ID)
	log.Debug().Msgf("Path = %v", e.Firstname)
}
