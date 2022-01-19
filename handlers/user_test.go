package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bananenpro/hbank-api/bindings"
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/db"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/router"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHandler_GetUsers(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
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
		tName         string
		user          *models.User
		pageSize      int
		exclude       string
		wantCode      int
		wantSuccess   bool
		wantAllInfo   bool
		wantUserCount int
	}{
		{tName: "All", user: user1, pageSize: 10, exclude: "", wantCode: http.StatusOK, wantSuccess: true, wantUserCount: 2},
		{tName: "Don't include self", user: user1, pageSize: 10, exclude: user1.Id.String(), wantCode: http.StatusOK, wantSuccess: true, wantUserCount: 1},
		{tName: "Only 1 user", user: user1, pageSize: 1, exclude: "", wantCode: http.StatusOK, wantSuccess: true, wantUserCount: 1},
		{tName: "Invalid page size", user: user1, pageSize: -1, exclude: "", wantCode: http.StatusBadRequest, wantSuccess: false},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?pageSize=%d&exclude=%s", tt.pageSize, tt.exclude), nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.Set("userId", tt.user.Id)

			err := handler.GetUsers(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))

			type usersResp struct {
				responses.Base
				Users []models.User `json:"users"`
			}
			var users usersResp
			json.Unmarshal(rec.Body.Bytes(), &users)
			assert.Equal(t, tt.wantUserCount, len(users.Users))
		})
	}
}

func TestHandler_GetUser(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
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
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"emailConfirmed":%t`, tt.user.EmailConfirmed))
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"twoFAOTPEnabled":%t`, tt.user.TwoFaOTPEnabled))
				} else {
					assert.NotContains(t, rec.Body.String(), "email")
					assert.NotContains(t, rec.Body.String(), "emailConfirmed")
					assert.NotContains(t, rec.Body.String(), "twoFAOTPEnabled")
				}
			}
		})
	}
}

func TestHandler_DeleteUser(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)
	gs := db.NewGroupStore(database)

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

	handler := New(us, gs)

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
			jsonBody := fmt.Sprintf(`{"password": "%s", "twoFAToken": "%s"}`, tt.password, tt.twoFAToken)
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

				cashLog, err := us.GetCashLog(tt.user, "", -1, -1, false)
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
}

func TestHandler_DeleteUserByDeleteToken(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)
	user1 := &models.User{
		Name:        "bob",
		Email:       "bob@gmail.com",
		DeleteToken: "123456",
	}
	us.Create(user1)

	user2 := &models.User{
		Name:        "paul",
		Email:       "paul@gmail.com",
		DeleteToken: "123456",
	}
	us.Create(user2)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		userId      string
		token       string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", userId: user1.Id.String(), token: "123456", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully deleted account"},
		{tName: "Wrong id", userId: uuid.NewString(), token: "123456", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong token", userId: user2.Id.String(), token: "654321", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/?token="+tt.token, nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.SetParamNames("id")
			c.SetParamValues(tt.userId)

			err := handler.DeleteUserByDeleteToken(c)

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
}

func TestHandler_UpdateUser(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:                    "bob",
		Email:                   "bob@gmail.com",
		DontSendInvitationEmail: true,
		ProfilePicturePrivacy:   "hi",
	}
	us.Create(user1)

	user2 := &models.User{
		Name:                    "bob2",
		Email:                   "bob2@gmail.com",
		DontSendInvitationEmail: true,
		ProfilePicturePrivacy:   "hi",
	}
	us.Create(user2)

	handler := New(us, nil)

	tests := []struct {
		tName                   string
		user                    *models.User
		dontSendInvitationEmail bool
		profilePicturePrivacy   string
		wantCode                int
		wantSuccess             bool
		wantMessage             string
	}{
		{tName: "Success", user: user1, dontSendInvitationEmail: false, profilePicturePrivacy: "everybody", wantCode: http.StatusOK, wantSuccess: true},
		{tName: "Invalid profilePicturePrivacy", user: user2, dontSendInvitationEmail: false, profilePicturePrivacy: "blablabla", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Invalid profile picture privacy"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"dontSendInvitationEmail": %t, "profilePicturePrivacy": "%s", "email": "bla@bla.bla", "password": "123456"}`, tt.dontSendInvitationEmail, tt.profilePicturePrivacy)
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.Set("userId", tt.user.Id)

			err := handler.UpdateUser(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			user, _ := us.GetById(tt.user.Id)
			if tt.wantSuccess {
				assert.Equal(t, tt.dontSendInvitationEmail, user.DontSendInvitationEmail)
				assert.Equal(t, tt.profilePicturePrivacy, user.ProfilePicturePrivacy)
			} else {
				assert.NotEqual(t, tt.dontSendInvitationEmail, user.DontSendInvitationEmail)
				assert.NotEqual(t, tt.profilePicturePrivacy, user.ProfilePicturePrivacy)
			}

			assert.Equal(t, tt.user.Email, user.Email)
			assert.Equal(t, tt.user.PasswordHash, user.PasswordHash)
		})
	}
}

func TestHandler_GetCurrentCash(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		CashLog: []models.CashLogEntry{
			{ChangeTitle: "Change1", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change2", Base: models.Base{Created: time.Now().Unix() + 100000}},
			{ChangeTitle: "Change3", Base: models.Base{Created: time.Now().Unix()}},
		},
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "peter",
		Email: "peter@gmail.com",
	}
	us.Create(user2)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		user        *models.User
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantTitle   string
	}{
		{tName: "Success", user: user1, wantCode: http.StatusOK, wantSuccess: true, wantTitle: "Change2"},
		{tName: "Empty cash log", user: user2, wantCode: http.StatusOK, wantSuccess: true, wantTitle: ""},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.Set("userId", tt.user.Id)

			err := handler.GetCurrentCash(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"title":"%s"`, tt.wantTitle))
			}
		})
	}
}

func TestHandler_GetCashLogEntryById(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		CashLog: []models.CashLogEntry{
			{ChangeTitle: "Change1", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change2", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change3", Base: models.Base{Created: time.Now().Unix()}},
		},
	}
	us.Create(user)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		entryId     uuid.UUID
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantTitle   string
	}{
		{tName: "Success", entryId: user.CashLog[2].Id, wantCode: http.StatusOK, wantSuccess: true, wantTitle: "Change3"},
		{tName: "Doesn't exist", entryId: uuid.New(), wantCode: http.StatusNotFound, wantSuccess: false, wantMessage: "Resource not found"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.Set("userId", user.Id)
			c.SetParamNames("id")
			c.SetParamValues(tt.entryId.String())

			err := handler.GetCashLogEntryById(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"title":"%s"`, tt.wantTitle))
			}
		})
	}
}

func TestHandler_GetCashLog(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		CashLog: []models.CashLogEntry{
			{ChangeTitle: "Change1", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change2", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change3", Base: models.Base{Created: time.Now().Unix()}},
		},
	}
	us.Create(user)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		pageSize    int
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantCount   int
	}{
		{tName: "Get all", pageSize: 10, wantCode: http.StatusOK, wantSuccess: true, wantCount: 3},
		{tName: "Get 1", pageSize: 1, wantCode: http.StatusOK, wantSuccess: true, wantCount: 1},
		{tName: "Invalid pageSize", pageSize: -1, wantCode: http.StatusBadRequest, wantSuccess: false, wantMessage: "Unsupported page size"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?pageSize=%d", tt.pageSize), nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.Set("userId", user.Id)
			c.SetParamNames("id")

			err := handler.GetCashLog(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				type cashLogResp struct {
					responses.Base
					CashLog []models.CashLogEntry `json:"log"`
				}
				var resp cashLogResp
				json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.Equal(t, tt.wantCount, len(resp.CashLog))
			}
		})
	}
}

func TestHandler_AddCashLogEntry(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "bob2",
		Email: "bob2@gmail.com",
	}
	us.Create(user2)

	handler := New(us, nil)

	tests := []struct {
		tName       string
		user        *models.User
		entry       bindings.AddCashLogEntry
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", user: user1, entry: bindings.AddCashLogEntry{Title: "Test"}, wantCode: http.StatusCreated, wantSuccess: true, wantMessage: "Successfully added new cash log entry"},
		{tName: "Title too short", user: user2, entry: bindings.AddCashLogEntry{Title: "    hi   "}, wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Title too short"},
		{tName: "Title too long", user: user2, entry: bindings.AddCashLogEntry{Title: "12345678901234567890123456789012"}, wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Title too long"},
		{tName: "Description too long", user: user2, entry: bindings.AddCashLogEntry{Title: "Test", Description: strings.Repeat("a", 257)}, wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Description too long"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.entry)
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(jsonBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.Set("userId", tt.user.Id)

			err := handler.AddCashLogEntry(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			log, _ := us.GetCashLog(tt.user, "", 0, 10, false)
			if tt.wantSuccess {
				assert.Equal(t, 1, len(log))
			} else {
				assert.Equal(t, 0, len(log))
			}
		})
	}
}
