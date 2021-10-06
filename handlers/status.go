package handlers

import (
	"net/http"

	"github.com/Bananenpro/hbank2-api/responses"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Status(c echo.Context) error {
	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "active",
	})
}
