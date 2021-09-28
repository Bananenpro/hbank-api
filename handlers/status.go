package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/responses"
)

func Status(c echo.Context) error {
	resp := responses.Status{
		Api: true,
	}
	return c.JSON(http.StatusOK, resp)
}
