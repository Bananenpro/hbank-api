package middlewares

import (
	"net/http"

	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/labstack/echo/v4"
)

func JWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authToken, err := c.Cookie("Auth-Token")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, responses.Base{
				Success: false,
				Message: "Missing Auth-Token Cookie",
			})
		}
		authTokenSignature, err := c.Cookie("Auth-Token-Signature")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, responses.Base{
				Success: false,
				Message: "Missing Auth-Token-Signature Cookie",
			})
		}

		userId, valid := services.VerifyAuthToken(authToken.Value + "." + authTokenSignature.Value)
		if !valid {
			return c.JSON(http.StatusUnauthorized, responses.Base{
				Success: false,
				Message: "Invalid JWT",
			})
		}

		c.Set("userId", userId)

		return next(c)
	}
}
