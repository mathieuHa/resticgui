package Models

import (
	//"errors"
	//"github.com/google/uuid"
	//"gorm.io/gorm"
	"time"
)

type Base struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (b *Base) IsValid() bool {
	return true
}
