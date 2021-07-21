package Models

type Permission struct {
	ID   string
	Name string `gorm:"unique"`
}
