package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	Id      uuid.UUID `gorm:"type:uuid;primaryKey"`
	Created int64     `gorm:"autoCreateTime"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	b.Id = uuid.New()
	return
}
