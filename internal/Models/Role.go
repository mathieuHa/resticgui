package Models

type Role struct {
	Base
	Name        string `gorm:"unique"`
	Description string
	ParentRole  string
	UserId      string
}
