package responses

import "github.com/Bananenpro/hbank-api/config"

type Config struct {
	CaptchaEnabled            bool   `json:"captchaEnabled"`
	EmailEnabled              bool   `json:"emailEnabled"`
	MinNameLength             int    `json:"minNameLength"`
	MaxNameLength             int    `json:"maxNameLength"`
	MinDescriptionLength      int    `json:"minDescriptionLength"`
	MaxDescriptionLength      int    `json:"maxDescriptionLength"`
	MinPasswordLength         int    `json:"minPasswordLength"`
	MaxPasswordLength         int    `json:"maxPasswordLength"`
	MinEmailLength            int    `json:"minEmailLength"`
	MaxEmailLength            int    `json:"maxEmailLength"`
	MaxProfilePictureFileSize int64  `json:"maxProfilePictureFileSize"`
	LoginTokenLifetime        int64  `json:"loginTokenLifetime"`
	EmailCodeLifetime         int64  `json:"emailCodeLifetime"`
	AuthTokenLifetime         int64  `json:"authTokenLifetime"`
	RefreshTokenLifetime      int64  `json:"refreshTokenLifetime"`
	SendEmailTimeout          int64  `json:"sendEmailTimeout"`
	MaxPageSize               int    `json:"maxPageSize"`
	IDProvider                string `json:"idProvider"`
}

type Status struct {
	Base
	Config Config `json:"config"`
}

func NewStatus() interface{} {
	return Status{
		Base: Base{
			Success: true,
			Message: "online",
		},
		Config: Config{
			CaptchaEnabled:            config.Data.CaptchaEnabled,
			EmailEnabled:              config.Data.EmailEnabled,
			MinNameLength:             config.Data.MinNameLength,
			MaxNameLength:             config.Data.MaxNameLength,
			MinDescriptionLength:      config.Data.MinDescriptionLength,
			MaxDescriptionLength:      config.Data.MaxDescriptionLength,
			MinPasswordLength:         config.Data.MinPasswordLength,
			MaxPasswordLength:         config.Data.MaxPasswordLength,
			MinEmailLength:            config.Data.MinEmailLength,
			MaxEmailLength:            config.Data.MaxEmailLength,
			MaxProfilePictureFileSize: config.Data.MaxProfilePictureFileSize,
			LoginTokenLifetime:        config.Data.LoginTokenLifetime,
			EmailCodeLifetime:         config.Data.EmailCodeLifetime,
			AuthTokenLifetime:         config.Data.AuthTokenLifetime,
			RefreshTokenLifetime:      config.Data.RefreshTokenLifetime,
			SendEmailTimeout:          config.Data.SendEmailTimeout,
			MaxPageSize:               config.Data.MaxPageSize,
			IDProvider:                config.Data.IDProvider,
		},
	}
}
