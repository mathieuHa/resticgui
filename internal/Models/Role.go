package Models

type Role struct {
	Base

	Name        string `json:"name" gorm:"unique"`
	Description string `json:"description"`
	ParentRole  uint   `json:"parentRole"`
	UserID      uint   `json:"userID"`
}
