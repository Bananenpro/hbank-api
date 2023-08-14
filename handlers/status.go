package handlers

import (
	"net/http"

	"github.com/juho05/h-bank/responses"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Status(c echo.Context) error {
	return c.JSON(http.StatusOK, responses.NewStatus())
}
