package models

import (
	"github.com/google/uuid"
)

type UserStore interface {
	GetById(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error

	GetEmailCode(user *User) (*EmailCode, error)
	DeleteEmailCode(code *EmailCode) error

	GetRefreshToken(user *User, code string) (*RefreshToken, error)
	RotateRefreshToken(user *User, oldRefreshToken *RefreshToken) (*RefreshToken, error)

	GetLoginTokenByCode(user *User, code string) (*LoginToken, error)
	GetLoginTokens(user *User) ([]LoginToken, error)
	DeleteLoginToken(token *LoginToken) error

	GetRecoveryCodeByCode(user *User, code string) (*RecoveryCode, error)
	GetRecoveryCodes(user *User) ([]RecoveryCode, error)
	NewRecoveryCodes(user *User) ([]RecoveryCode, error)

	GetConfirmEmailLastSent(email string) (int64, error)
	SetConfirmEmailLastSent(email string, time int64) error
}

type User struct {
	Base
	Name             string
	Email            string `gorm:"unique"`
	PasswordHash     []byte
	ProfilePicture   []byte
	ProfilePictureId uuid.UUID `gorm:"type:uuid"`
	EmailConfirmed   bool
	TwoFaOTPEnabled  bool
	OtpSecret        string
	OtpQrCode        []byte
	EmailCode        EmailCode      `gorm:"constraint:OnDelete:CASCADE"`
	RefreshTokens    []RefreshToken `gorm:"constraint:OnDelete:CASCADE"`
	LoginTokens      []LoginToken   `gorm:"constraint:OnDelete:CASCADE"`
	RecoveryCodes    []RecoveryCode `gorm:"constraint:OnDelete:CASCADE"`
}

type ConfirmEmailLastSent struct {
	Base
	Email    string `gorm:"unique"`
	LastSent int64
}

type EmailCode struct {
	Base
	Code           string
	ExpirationTime int64
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type RefreshToken struct {
	Base
	Code           string
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
