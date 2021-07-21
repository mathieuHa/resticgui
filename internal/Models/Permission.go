package Models

type Permission struct {
	ID   string `gorm:"primaryKey"`
	Name string `gorm:"unique"`
}
