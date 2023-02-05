package responses

import "github.com/Bananenpro/hbank-api/config"

type Config struct {
	EmailEnabled              bool   `json:"emailEnabled"`
	MinNameLength             int    `json:"minNameLength"`
	MaxNameLength             int    `json:"maxNameLength"`
	MinDescriptionLength      int    `json:"minDescriptionLength"`
	MaxDescriptionLength      int    `json:"maxDescriptionLength"`
	MaxProfilePictureFileSize int64  `json:"maxProfilePictureFileSize"`
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
			EmailEnabled:              config.Data.EmailEnabled,
			MinNameLength:             config.Data.MinNameLength,
			MaxNameLength:             config.Data.MaxNameLength,
			MinDescriptionLength:      config.Data.MinDescriptionLength,
			MaxDescriptionLength:      config.Data.MaxDescriptionLength,
			MaxProfilePictureFileSize: config.Data.MaxProfilePictureFileSize,
			MaxPageSize:               config.Data.MaxPageSize,
			IDProvider:                config.Data.IDProvider,
		},
	}
}
