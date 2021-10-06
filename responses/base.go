package responses

import "gitlab.com/Bananenpro05/hbank2-api/config"

type Base struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewUnexpectedError(err error) Base {
	if config.Data.Debug {
		return Base{
			Success: false,
			Message: "Error: " + err.Error(),
		}
	} else {
		return Base{
			Success: false,
			Message: "Due to an unexpected error the user couldn't be registered",
		}
	}
}
