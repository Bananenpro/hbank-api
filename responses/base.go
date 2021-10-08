package responses

import "github.com/Bananenpro/hbank2-api/config"

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
			Message: "An unexpected error occurred",
		}
	}
}

func NewNotFound() Base {
	return Base{
		Success: false,
		Message: "Resource not found",
	}
}
