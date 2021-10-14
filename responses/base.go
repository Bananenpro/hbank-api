package responses

import (
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/services"
)

type Base struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func New(success bool, message string, lang string) Base {
	return Base{
		Success: success,
		Message: services.Tr(message, lang),
	}
}

func NewUnexpectedError(err error, lang string) Base {
	if config.Data.Debug {
		return New(false, "Error: "+err.Error(), lang)
	} else {
		return New(false, "An unexpected error occured", lang)
	}
}

func NewNotFound(lang string) Base {
	return New(false, "Resource not found", lang)
}

func NewInvalidRequestBody(lang string) Base {
	return New(false, "Invalid request body", lang)
}
