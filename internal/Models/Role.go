package Models

type Role struct {
	ID          string
	Name        string `gorm:"unique"`
	Description string
	ParentRole  string
	UserId      string
}
