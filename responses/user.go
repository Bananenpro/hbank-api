package responses

import "github.com/Bananenpro/hbank-api/models"

type AuthUser struct {
	Id                      string `json:"id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	EmailConfirmed          bool   `json:"emailConfirmed"`
	TwoFAOTPEnabled         bool   `json:"twoFAOTPEnabled"`
	ProfilePictureId        string `json:"profilePictureId"`
	ProfilePicturePrivacy   string `json:"profilePicturePrivacy"`
	DontSendInvitationEmail bool   `json:"dontSendInvitationEmail"`
	DeleteToken             string `json:"deleteToken,omitempty"`
}

type User struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	ProfilePictureId string `json:"profilePictureId"`
}

type CashLogEntryDetailed struct {
	Id          string `json:"id"`
	Time        int64  `json:"time"`
	Title       string `json:"title"`
	Description string `json:"description"`

	Ct1    int `json:"ct1"`
	Ct2    int `json:"ct2"`
	Ct5    int `json:"ct5"`
	Ct10   int `json:"ct10"`
	Ct20   int `json:"ct20"`
	Ct50   int `json:"ct50"`
	Eur1   int `json:"eur1"`
	Eur2   int `json:"eur2"`
	Eur5   int `json:"eur5"`
	Eur10  int `json:"eur10"`
	Eur20  int `json:"eur20"`
	Eur50  int `json:"eur50"`
	Eur100 int `json:"eur100"`
	Eur200 int `json:"eur200"`
	Eur500 int `json:"eur500"`

	Amount     int `json:"amount"`
	Difference int `json:"difference"`
}

type CashLogEntry struct {
	Id         string `json:"id"`
	Time       int64  `json:"time"`
	Title      string `json:"title"`
	Amount     int    `json:"amount"`
	Difference int    `json:"difference"`
}

func NewCashLogEntry(entry *models.CashLogEntry) interface{} {
	type cashLogEntryResp struct {
		Base
		CashLogEntryDetailed
	}
	return cashLogEntryResp{
		Base: Base{
			Success: true,
		},
		CashLogEntryDetailed: CashLogEntryDetailed{
			Id:          entry.Id.String(),
			Time:        entry.Created,
			Title:       entry.ChangeTitle,
			Description: entry.ChangeDescription,

			Ct1:    entry.Ct1,
			Ct2:    entry.Ct2,
			Ct5:    entry.Ct5,
			Ct10:   entry.Ct10,
			Ct20:   entry.Ct20,
			Ct50:   entry.Ct50,
			Eur1:   entry.Eur1,
			Eur2:   entry.Eur2,
			Eur5:   entry.Eur5,
			Eur10:  entry.Eur10,
			Eur20:  entry.Eur20,
			Eur50:  entry.Eur50,
			Eur100: entry.Eur100,
			Eur200: entry.Eur200,
			Eur500: entry.Eur500,

			Amount:     entry.TotalAmount,
			Difference: entry.ChangeDifference,
		},
	}
}

func NewCashLog(log []models.CashLogEntry, count int64) interface{} {
	type cashLogResp struct {
		Base
		Count   int64          `json:"count"`
		CashLog []CashLogEntry `json:"log"`
	}

	entries := make([]CashLogEntry, len(log))

	for i, entry := range log {
		entries[i] = CashLogEntry{
			Id:    entry.Id.String(),
			Time:  entry.Created,
			Title: entry.ChangeTitle,

			Amount:     entry.TotalAmount,
			Difference: entry.ChangeDifference,
		}
	}

	return cashLogResp{
		Base: Base{
			Success: true,
		},
		Count:   count,
		CashLog: entries,
	}
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
			Id:                      user.Id.String(),
			Name:                    user.Name,
			Email:                   user.Email,
			EmailConfirmed:          user.EmailConfirmed,
			TwoFAOTPEnabled:         user.TwoFaOTPEnabled,
			ProfilePictureId:        user.ProfilePictureId.String(),
			ProfilePicturePrivacy:   user.ProfilePicturePrivacy,
			DontSendInvitationEmail: user.DontSendInvitationEmail,
			DeleteToken:             user.DeleteToken,
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
			Id:               user.Id.String(),
			Name:             user.Name,
			ProfilePictureId: user.ProfilePictureId.String(),
		},
	}
}

func NewUsers(users []models.User, count int64) interface{} {
	userDTOs := make([]User, len(users))
	for i, u := range users {
		userDTOs[i].Id = u.Id.String()
		userDTOs[i].Name = u.Name
		userDTOs[i].ProfilePictureId = u.ProfilePictureId.String()
	}

	type usersResp struct {
		Base
		Count int64  `json:"count"`
		Users []User `json:"users"`
	}

	return usersResp{
		Base: Base{
			Success: true,
		},
		Count: count,
		Users: userDTOs,
	}
}
