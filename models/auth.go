package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailCode struct {
	Base
	Code           string
	ExpirationTime int64
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type RefreshToken struct {
	Base
	Code           string
	DeviceId       uuid.UUID `gorm:"type:uuid"`
	ExpirationTime int64
	Used           bool
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type LoginToken struct {
	Base
	Code           string
	ExpirationTime int64
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type RecoveryCode struct {
	Base
	Code   string
	UserId uuid.UUID `gorm:"type:uuid"`
}

func authAutoMigrate(db *gorm.DB) (errs []error) {
	err := db.AutoMigrate(EmailCode{})
	if err != nil {
		errs = append(errs, err)
	}

	err = db.AutoMigrate(RefreshToken{})
	if err != nil {
		errs = append(errs, err)
	}

	err = db.AutoMigrate(LoginToken{})
	if err != nil {
		errs = append(errs, err)
	}

	err = db.AutoMigrate(RecoveryCode{})
	if err != nil {
		errs = append(errs, err)
	}

	return
}
