package Models

import "gorm.io/gorm"

//import "github.com/google/uuid"

type Role struct {
	gorm.Model

	Name        string `json:"name" gorm:"unique"`
	Description string `json:"description"`
	ParentRole  uint   `json:"parent_role"`
	UserID      uint   //uuid.UUID `json:"user_id"`
}
