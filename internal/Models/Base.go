package Models

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type Base struct {
	ID        uuid.UUID `sql:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *Base) BeforeCreate() (err error) {
	b.ID = uuid.New()
	if !b.IsValid() {
		err = errors.New("can't save invalid data")
	}
	return
}

func (b *Base) IsValid() bool {
	return true
}
