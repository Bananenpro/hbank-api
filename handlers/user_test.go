package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/db"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/router"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHandler_GetUser(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:            "bob",
		Email:           "bob@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
	}
	us.Create(user1)

	user2 := &models.User{
		Name:            "peter",
		Email:           "peter@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
	}
	us.Create(user2)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		user        *models.User
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantAllInfo bool
	}{
		{tName: "Doesn't exist", wantCode: http.StatusNotFound, wantSuccess: false, wantMessage: "Resource not found"},
		{tName: "Auth user", user: user1, wantCode: http.StatusOK, wantSuccess: true, wantAllInfo: true},
		{tName: "Not auth user", user: user2, wantCode: http.StatusOK, wantSuccess: true, wantAllInfo: false},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.SetParamNames("id")
			if tt.user != nil {
				c.SetParamValues(tt.user.Id.String())
			} else {
				c.SetParamValues(uuid.NewString())
			}
			c.Set("userId", user1.Id)

			err := handler.GetUser(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"id":"%s"`, tt.user.Id.String()))
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"name":"%s"`, tt.user.Name))

				if tt.wantAllInfo {
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"email":"%s"`, tt.user.Email))
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"email_confirmed":%t`, tt.user.EmailConfirmed))
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"two_fa_otp_enabled":%t`, tt.user.TwoFaOTPEnabled))
				} else {
					assert.NotContains(t, rec.Body.String(), fmt.Sprintf(`"email"`))
					assert.NotContains(t, rec.Body.String(), fmt.Sprintf(`"email_confirmed"`))
					assert.NotContains(t, rec.Body.String(), fmt.Sprintf(`"two_fa_otp_enabled"`))
				}
			}
		})
	}

	db.DeleteTestDB()
}

func TestHandler_DeleteUser(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	password := "123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), config.Data.BcryptCost)
	user1 := &models.User{
		Name:              "bob",
		Email:             "bob@gmail.com",
		PasswordHash:      hash,
		ConfirmEmailCode:  models.ConfirmEmailCode{},
		ResetPasswordCode: models.ResetPasswordCode{},
		ChangeEmailCode:   models.ChangeEmailCode{},
		RefreshTokens: []models.RefreshToken{
			{CodeHash: services.HashToken("abcde")},
			{CodeHash: services.HashToken("edcba")},
		},
		PasswordTokens: []models.PasswordToken{
			{CodeHash: services.HashToken("abcde")},
			{CodeHash: services.HashToken("edcba")},
		},
		TwoFATokens: []models.TwoFAToken{
			{CodeHash: services.HashToken("1234567890"), ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime},
			{CodeHash: services.HashToken("12345678901"), ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime},
		},
		RecoveryCodes: []models.RecoveryCode{
			{CodeHash: services.HashToken("abcde")},
			{CodeHash: services.HashToken("edcba")},
		},
		CashLog: []models.CashLogEntry{
			{ChangeTitle: "Hello"},
			{ChangeTitle: "Hello2"},
		},
	}
	us.Create(user1)
	us.SetConfirmEmailLastSent(user1.Email, time.Now().Unix())
	us.SetForgotPasswordEmailLastSent(user1.Email, time.Now().Unix())

	user2 := &models.User{
		Name:         "bob2",
		Email:        "bob2@gmail.com",
		PasswordHash: hash,
		TwoFATokens:  []models.TwoFAToken{{CodeHash: services.HashToken("1234567890"), ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
	}
	us.Create(user2)

	user3 := &models.User{
		Name:         "bob3",
		Email:        "bob3@gmail.com",
		PasswordHash: hash,
		TwoFATokens:  []models.TwoFAToken{{CodeHash: services.HashToken("1234567890"), ExpirationTime: 0}},
	}
	us.Create(user3)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		user        *models.User
		twoFAToken  string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", user: user1, twoFAToken: "1234567890", password: "123456", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully deleted account"},
		{tName: "Wrong password", user: user2, twoFAToken: "1234567890", password: "654321", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong two factor token", user: user2, twoFAToken: "0987654321", password: "123456", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Expired two factor token", user: user3, twoFAToken: "1234567890", password: "123456", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"password": "%s", "two_fa_token": "%s"}`, tt.password, tt.twoFAToken)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.Set("userId", tt.user.Id)

			err := handler.DeleteUser(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			user, _ := us.GetById(tt.user.Id)
			if tt.wantSuccess {
				assert.Nil(t, user)

				confirmEmailCode, err := us.GetConfirmEmailCode(tt.user)
				assert.NoError(t, err)
				assert.Nil(t, confirmEmailCode, "ConfirmEmailCode")

				resetPasswordCode, err := us.GetResetPasswordCode(tt.user)
				assert.NoError(t, err)
				assert.Nil(t, resetPasswordCode, "ResetPasswordCode")

				changeEmailCode, err := us.GetChangeEmailCode(tt.user)
				assert.NoError(t, err)
				assert.Nil(t, changeEmailCode, "ChangeEmailCode")

				refreshTokens, err := us.GetRefreshTokens(tt.user)
				assert.NoError(t, err)
				assert.Empty(t, refreshTokens, "RefreshTokens")

				passwordTokens, err := us.GetPasswordTokens(tt.user)
				assert.NoError(t, err)
				assert.Empty(t, passwordTokens, "PasswordTokens")

				twoFATokens, err := us.GetTwoFATokens(tt.user)
				assert.NoError(t, err)
				assert.Empty(t, twoFATokens, "TwoFATokens")

				recoveryCodes, err := us.GetRecoveryCodes(tt.user)
				assert.NoError(t, err)
				assert.Empty(t, recoveryCodes, "RecoveryCodes")

				cashLog, err := us.GetCashLog(tt.user, -1, -1, false)
				assert.NoError(t, err)
				assert.Empty(t, cashLog, "CashLog")

				confirmEmailLastSent, err := us.GetConfirmEmailLastSent(tt.user.Email)
				assert.NoError(t, err)
				assert.Zero(t, confirmEmailLastSent, "ConfirmEmailLastSent")

				forgotPasswordEmailLastSent, err := us.GetForgotPasswordEmailLastSent(tt.user.Email)
				assert.NoError(t, err)
				assert.Zero(t, forgotPasswordEmailLastSent, "ForgotPasswordEmailLastSent")
			} else {
				assert.NotNil(t, user)
			}

			if user != nil && tt.twoFAToken == "1234567890" && tt.password == "123456" {
				tokens, _ := us.GetTwoFATokens(user)
				assert.Empty(t, tokens)
			}
		})
	}

	db.DeleteTestDB()
}

func TestHandler_DeleteUserByConfirmEmailCode(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)
	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		ConfirmEmailCode: models.ConfirmEmailCode{
			CodeHash: services.HashToken("123456"),
		},
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "paul",
		Email: "paul@gmail.com",
		ConfirmEmailCode: models.ConfirmEmailCode{
			CodeHash: services.HashToken("123456"),
		},
		EmailConfirmed: true,
	}
	us.Create(user2)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		userId      string
		code        string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", userId: user1.Id.String(), code: "123456", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully deleted account"},
		{tName: "Wrong id", userId: uuid.NewString(), code: "123456", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong code", userId: user2.Id.String(), code: "654321", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Already confirmed", userId: user2.Id.String(), code: "123456", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"code": "%s"}`, tt.code)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.SetParamNames("id")
			c.SetParamValues(tt.userId)

			err := handler.DeleteUserByConfirmEmailCode(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			userId, _ := uuid.Parse(tt.userId)
			user, _ := us.GetById(userId)
			if tt.wantSuccess {
				assert.Nil(t, user)
			} else if tt.tName != "Wrong id" {
				assert.NotNil(t, user)
			}
		})
	}

	db.DeleteTestDB()
}
