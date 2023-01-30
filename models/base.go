package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	Id      string `gorm:"primaryKey"`
	Created int64  `gorm:"autoCreateTime"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.Id == "" {
		b.Id = uuid.NewString()
	}
	return
}
