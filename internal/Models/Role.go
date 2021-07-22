package Models

type Role struct {
	Base
	Name        string `json:"name" gorm:"unique"`
	Description string `json:"description"`
	ParentRole  string `json:"parent_role"`
	UserId      string `json:"user_id"`
}
