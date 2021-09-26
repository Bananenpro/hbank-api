package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/responses"
)

func HealthCheck(c echo.Context) error {
	resp := responses.HealthCheck{
		ApiActive: true,
	}
	return c.JSON(http.StatusOK, resp)
}
