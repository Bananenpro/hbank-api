package handlers

import (
	"bytes"
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/Bananenpro/hbank-api/bindings"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// /v1/user?includeSelf=bool (GET)
func (h *Handler) GetUsers(c echo.Context) error {
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists())
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
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists())
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter"))
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

// /v1/user (DELETE)
func (h *Handler) DeleteUser(c echo.Context) error {
	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists())
	}

	var body bindings.DeleteUser
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody())
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}

	twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if twoFAToken == nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}
	h.userStore.DeleteTwoFAToken(twoFAToken)
	if twoFAToken.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}

	h.userStore.Delete(user)
	return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted account"))
}

// /v1/user/:email (DELETE)
func (h *Handler) DeleteUserByConfirmEmailCode(c echo.Context) error {
	var body bindings.DeleteUserByConfirmEmailCode
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody())
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter"))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	code, err := h.userStore.GetConfirmEmailCode(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	if subtle.ConstantTimeCompare(code.CodeHash, services.HashToken(body.ConfirmEmailCode)) == 1 {
		if !user.EmailConfirmed {
			h.userStore.Delete(user)
			return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted account"))
		}
		h.userStore.DeleteConfirmEmailCode(code)
	}
	return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
}
