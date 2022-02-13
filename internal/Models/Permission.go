package Models

type Permission struct {
	Base

	Name string `json:"name" gorm:"unique"`
}
