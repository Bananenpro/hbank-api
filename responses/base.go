package responses

import (
	"net/http"
	"strings"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/labstack/echo/v4"
)

type Base struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Id struct {
	Base
	Id string `json:"id"`
}

func New(success bool, message string, lang string) Base {
	return Base{
		Success: success,
		Message: services.Tr(message, lang),
	}
}

func HandleHTTPError(err error, c echo.Context) {
	headerValues := c.Request().Header["Accept-Language"]
	headerValue := strings.Join(headerValues, ",")
	lang := services.GetLanguageFromAcceptLanguageHeader(headerValue)

	code := http.StatusInternalServerError
	message := NewUnexpectedError(err, "").Message
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	c.JSON(code, New(false, message, lang))

	if code != http.StatusNotFound && code != http.StatusMethodNotAllowed {
		c.Logger().Error(err)
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
