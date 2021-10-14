package responses

import "github.com/Bananenpro/hbank-api/config"

type Base struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func New(success bool, message string) Base {
	return Base{
		Success: success,
		Message: message,
	}
}

func NewUnexpectedError(err error) Base {
	if config.Data.Debug {
		return New(false, "Error: "+err.Error())
	} else {
		return New(false, "An unexpected error occured")
	}
}

func NewNotFound() Base {
	return New(false, "Resource not found")
}

func NewInvalidRequestBody() Base {
	return New(false, "Invalid request body")
}
