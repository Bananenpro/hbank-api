package middlewares

import (
	"strings"

	"github.com/juho05/h-bank/services"
	"github.com/labstack/echo/v4"
)

func Lang(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		headerValues := c.Request().Header["Accept-Language"]
		headerValue := strings.Join(headerValues, ",")
		c.Set("lang", services.GetLanguageFromAcceptLanguageHeader(headerValue))
		return next(c)
	}
}
