package Models

import (
	"time"
)

type User struct {
	Base

	Firstname          string     `json:"firstname" binding:"required,min=3"`
	Lastname           string     `json:"lastname" binding:"required,min=3"`
	Email              string     `json:"email" gorm:"unique" binding:"required,email"`
	Password           string     `json:"password"`
	PasswordLastSet    *time.Time `json:"password_last_set"`
	PasswordFailCount  uint8      `json:"password_fail_count"`
	ResetPassword      bool       `json:"reset_password"`
	LastConnectionDate time.Time  `json:"last_connection_date"`
	Mfa                bool       `json:"mfa"`
	MfaToken           string     `json:"mfa_token"`
	Role               Role
	Active             bool `json:"active"`
}
