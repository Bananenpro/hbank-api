package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/responses"
)

func (h *Handler) Status(c echo.Context) error {
	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "active",
	})
}
