package Models

type Role struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"unique"`
	Description string
	ParentRole  string
	UserId      string
}
