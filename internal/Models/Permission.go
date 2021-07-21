package Models

type Permission struct {
	Base
	Name string `gorm:"unique"`
}
