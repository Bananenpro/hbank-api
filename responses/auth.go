package responses

import "gitlab.com/Bananenpro05/hbank2-api/config"

type RegisterSuccess struct {
	Base
	UserId    string `json:"user_id"`
	UserEmail string `json:"user_email"`
}

type RegisterInvalid struct {
	Base
	MinNameLength     int `json:"min_name_length"`
	MinPasswordLength int `json:"min_password_length"`
	MaxNameLength     int `json:"max_name_length"`
	MaxPasswordLength int `json:"max_password_length"`
}

type Login struct {
	Base
	LoginToken string `json:"login_token"`
}

func NewRegisterInvalid(message string) RegisterInvalid {
	return RegisterInvalid{
		Base: Base{
			Success: false,
			Message: message,
		},
		MinNameLength:     config.Data.UserMinNameLength,
		MinPasswordLength: config.Data.UserMinPasswordLength,
		MaxNameLength:     config.Data.UserMaxNameLength,
		MaxPasswordLength: config.Data.UserMaxPasswordLength,
	}
}

func NewInvalidCredentials() Base {
	return Base{
		Success: false,
		Message: "Invalid credentials",
	}
}
