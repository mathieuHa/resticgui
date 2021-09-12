package Models

import (
	//"errors"
	//"github.com/google/uuid"
	//"gorm.io/gorm"
	"time"
)

type Base struct {
	//	ID        uuid.UUID `json:"id" sql:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// func (b *Base) BeforeCreate(*gorm.DB) (err error) {
// 	b.ID = uuid.New()
// 	if !b.IsValid() {
// 		err = errors.New("can't save invalid data")
// 	}
// 	return
// }

func (b *Base) IsValid() bool {
	return true
}
