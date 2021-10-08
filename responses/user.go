package responses

import "github.com/Bananenpro/hbank2-api/models"

type AuthUser struct {
	Base
	Id              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	EmailConfirmed  bool   `json:"email_confirmed"`
	TwoFAOTPEnabled bool   `json:"two_fa_otp_enabled"`
}

type User struct {
	Base
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewAuthUser(user *models.User) AuthUser {
	return AuthUser{
		Base: Base{
			Success: true,
		},
		Id:              user.Id.String(),
		Name:            user.Name,
		Email:           user.Email,
		EmailConfirmed:  user.EmailConfirmed,
		TwoFAOTPEnabled: user.TwoFaOTPEnabled,
	}
}

func NewUser(user *models.User) User {
	return User{
		Base: Base{
			Success: true,
		},
		Id:   user.Id.String(),
		Name: user.Name,
	}
}
