package Models

import "github.com/google/uuid"

type Role struct {
	Base
	Name        string `json:"name" gorm:"unique"`
	Description string `json:"description"`
	ParentRole  uuid.UUID `json:"parent_role"`
	UserID      uuid.UUID `json:"user_id"`
}
