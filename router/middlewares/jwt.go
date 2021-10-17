package middlewares

import (
	"net/http"

	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/labstack/echo/v4"
)

func JWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		lang := c.Get("lang").(string)
		authToken, err := c.Cookie("Auth-Token")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, responses.New(false, "Missing Auth-Token Cookie", lang))
		}
		authTokenSignature, err := c.Cookie("Auth-Token-Signature")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, responses.New(false, "Missing Auth-Token-Signature Cookie", lang))
		}

		userId, valid := services.VerifyAuthToken(authToken.Value + "." + authTokenSignature.Value)
		if !valid {
			return c.JSON(http.StatusUnauthorized, responses.New(false, "Invalid JWT", lang))
		}

		c.Set("userId", userId)

		return next(c)
	}
}
