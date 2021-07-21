package Models

import (
	"time"
)

type User struct {
	ID                 string
	Firstname          string
	Lastname           string
	Email              string `gorm:"unique"`
	Password           string
	PasswordLastSet    *time.Time
	PasswordFailCount  uint8
	ResetPassword      bool
	RegisteredDate     time.Time
	LastConnectionDate time.Time
	Mfa                bool
	MfaToken           string
	Role               Role
}
