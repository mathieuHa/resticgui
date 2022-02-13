package Models

type Role struct {
	Base

	Name        string `json:"name" gorm:"unique"`
	Description string `json:"description"`
	ParentRole  uint   `json:"parent_role"`
	UserID      uint   //uuid.UUID `json:"user_id"`
}
