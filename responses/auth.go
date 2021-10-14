package responses

import (
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/services"
)

type RegisterSuccess struct {
	Base
	UserId    string   `json:"user_id"`
	UserEmail string   `json:"user_email"`
	Codes     []string `json:"recovery_codes"`
}

type RegisterInvalid struct {
	Base
	MinNameLength     int `json:"min_name_length"`
	MinPasswordLength int `json:"min_password_length"`
	MaxNameLength     int `json:"max_name_length"`
	MaxPasswordLength int `json:"max_password_length"`
}

type Token struct {
	Base
	Token string `json:"token"`
}

type RecoveryCodes struct {
	Base
	Codes []string `json:"recovery_codes"`
}

func NewRegisterInvalid(message string, lang string) RegisterInvalid {
	return RegisterInvalid{
		Base: Base{
			Success: false,
			Message: services.Tr(message, lang),
		},
		MinNameLength:     config.Data.UserMinNameLength,
		MinPasswordLength: config.Data.UserMinPasswordLength,
		MaxNameLength:     config.Data.UserMaxNameLength,
		MaxPasswordLength: config.Data.UserMaxPasswordLength,
	}
}

func NewInvalidCredentials(lang string) Base {
	return New(false, "Invalid credentials", lang)
}

func NewUserNoLongerExists(lang string) Base {
	return New(false, "The user does no longer exist", lang)
}
