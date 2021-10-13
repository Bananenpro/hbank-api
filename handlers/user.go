package handlers

import (
	"bytes"
	"net/http"

	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// /v1/user?includeSelf=bool (GET)
func (h *Handler) GetUsers(c echo.Context) error {
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil || authUser == nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	var users []models.User
	if services.StrToBool(c.QueryParams().Get("includeSelf")) {
		users, err = h.userStore.GetAll(nil)
	} else {
		users, err = h.userStore.GetAll(authUser)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusOK, responses.NewUsers(users))
}

// /v1/user/:id (GET)
func (h *Handler) GetUser(c echo.Context) error {
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil || authUser == nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid or missing id parameter",
		})
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound())
	}

	if bytes.Equal(authUserId[:], userId[:]) {
		return c.JSON(http.StatusOK, responses.NewAuthUser(authUser))
	}
	return c.JSON(http.StatusOK, responses.NewUser(user))
}
