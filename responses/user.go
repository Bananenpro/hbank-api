package responses

import "github.com/Bananenpro/hbank-api/models"

type AuthUser struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	EmailConfirmed  bool   `json:"email_confirmed"`
	TwoFAOTPEnabled bool   `json:"two_fa_otp_enabled"`
}

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewAuthUser(user *models.User) interface{} {
	type authUserResp struct {
		Base
		AuthUser
	}
	return authUserResp{
		Base: Base{
			Success: true,
		},
		AuthUser: AuthUser{
			Id:              user.Id.String(),
			Name:            user.Name,
			Email:           user.Email,
			EmailConfirmed:  user.EmailConfirmed,
			TwoFAOTPEnabled: user.TwoFaOTPEnabled,
		},
	}
}

func NewUser(user *models.User) interface{} {
	type userResp struct {
		Base
		User
	}
	return userResp{
		Base: Base{
			Success: true,
		},
		User: User{
			Id:   user.Id.String(),
			Name: user.Name,
		},
	}
}

func NewUsers(users []models.User) interface{} {
	userDTOs := make([]User, len(users))
	for i, u := range users {
		userDTOs[i].Id = u.Id.String()
		userDTOs[i].Name = u.Name
	}

	type usersResp struct {
		Base
		Users []User
	}

	return usersResp{
		Base: Base{
			Success: true,
		},
		Users: userDTOs,
	}
}
