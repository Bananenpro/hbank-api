package handlers

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bananenpro/hbank2-api/config"
	"github.com/Bananenpro/hbank2-api/db"
	"github.com/Bananenpro/hbank2-api/models"
	"github.com/Bananenpro/hbank2-api/router"
	"github.com/labstack/echo/v4"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewInMemory()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}
	db.Clear(database)

	us := db.NewUserStore(database)

	us.Create(&models.User{
		Name:  "bob",
		Email: "exists@gmail.com",
	})

	handler := New(us)

	tests := []struct {
		tName       string
		name        string
		email       string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Successful register", name: "bob", email: "bob@gmail.com", password: "123456", wantCode: http.StatusCreated, wantSuccess: true, wantMessage: "Successfully registered new user"},
		{tName: "User does already exist", name: "bob", email: "exists@gmail.com", password: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "The user with this email does already exist"},
		{tName: "Name too short", name: strings.Repeat("a", config.Data.UserMinNameLength-1), email: "bob@gmail.com", password: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Name too short"},
		{tName: "Name too long", name: strings.Repeat("a", config.Data.UserMaxNameLength+1), email: "bob@gmail.com", password: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Name too long"},
		{tName: "Password too short", name: "bob", email: "bob@gmail.com", password: strings.Repeat("a", config.Data.UserMinPasswordLength-1), wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Password too short"},
		{tName: "Password too long", name: "bob", email: "bob@gmail.com", password: strings.Repeat("a", config.Data.UserMaxPasswordLength+1), wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Password too long"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"name":"%s","email": "%s","password":"%s"}`, tt.name, tt.email, tt.password)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.Register(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if rec.Code == http.StatusCreated {
				us.DeleteByEmail(tt.email)
			}
		})
	}
}

func TestHandler_SendConfirmEmail(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewInMemory()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}
	db.Clear(database)

	us := db.NewUserStore(database)

	err = us.Create(&models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
	})

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Valid request", email: "bob@gmail.com", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "If the email address is linked to a user whose email has not yet been confirmed, a code has been sent to the specified address"},
		{tName: "Invalid request", email: "hehehe", wantCode: http.StatusBadRequest, wantSuccess: false, wantMessage: "Missing or invalid email parameter"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.SetParamNames("email")
			c.SetParamValues(tt.email)

			err := handler.SendConfirmEmail(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				user, err := us.GetByEmail(tt.email)
				assert.NoError(t, err)
				assert.NotNil(t, user)

				if user != nil {
					emailCode, err := us.GetEmailCode(user)
					assert.NoError(t, err)
					assert.NotNil(t, emailCode)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					c := r.NewContext(req, rec)
					c.SetParamNames("email")
					c.SetParamValues(tt.email)

					err = handler.SendConfirmEmail(c)

					assert.NoError(t, err)
					assert.Equal(t, http.StatusTooManyRequests, rec.Code)
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, false))
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, "Please wait at least 2 minutes between confirm email requests"))
				}
			}
		})
	}
}

func TestHandler_VerifyConfirmEmailCode(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewInMemory()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}
	db.Clear(database)

	us := db.NewUserStore(database)

	us.Create(&models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		EmailCode: models.EmailCode{
			Code:           "123456",
			ExpirationTime: time.Now().UnixMilli() + config.Data.EmailCodeLifetime,
		},
	})

	us.Create(&models.User{
		Name:  "paul",
		Email: "paul@gmail.com",
		EmailCode: models.EmailCode{
			Code:           "123456",
			ExpirationTime: time.Now().UnixMilli() + config.Data.EmailCodeLifetime,
		},
	})

	us.Create(&models.User{
		Name:  "peter",
		Email: "peter@gmail.com",
		EmailCode: models.EmailCode{
			Code:           "123456",
			ExpirationTime: 0,
		},
	})

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		code        string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", code: "123456", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully confirmed email address"},
		{tName: "Wrong email", email: "paula@gmail.com", code: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Email was not confirmed"},
		{tName: "Wrong code", email: "paul@gmail.com", code: "654321", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Email was not confirmed"},
		{tName: "Expired code", email: "peter@gmail.com", code: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Email was not confirmed"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "code": "%s"}`, tt.email, tt.code)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.VerifyConfirmEmailCode(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			user, err := us.GetByEmail(tt.email)
			if user != nil {
				code, err := us.GetEmailCode(user)
				assert.NoError(t, err)
				assert.Equal(t, tt.code == "123456", code == nil, "Code was (not) deleted from database")
				assert.Equal(t, tt.wantSuccess, user.EmailConfirmed, "Email (not) confirmed")
			}

		})
	}
}

func TestHandler_Activate2FAOTP(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewInMemory()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}
	db.Clear(database)

	us := db.NewUserStore(database)

	password, err := bcrypt.GenerateFromPassword([]byte("password"), config.Data.BcryptCost)
	us.Create(&models.User{
		Name:         "bob",
		Email:        "bob@gmail.com",
		PasswordHash: password,
	})

	us.Create(&models.User{
		Name:            "paul",
		Email:           "paul@gmail.com",
		PasswordHash:    password,
		TwoFaOTPEnabled: true,
	})

	us.Create(&models.User{
		Name:            "peter",
		Email:           "peter@gmail.com",
		PasswordHash:    password,
		TwoFaOTPEnabled: true,
	})

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", password: "password", wantCode: http.StatusOK, wantSuccess: true},
		{tName: "Already activated", email: "paul@gmail.com", password: "password", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "TwoFaOTP is already activated"},
		{tName: "Wrong email", email: "retep@gmail.com", password: "password", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong password", email: "peter@gmail.com", password: "drowssap", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, tt.email, tt.password)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.Activate2FAOTP(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)

			if tt.wantSuccess {
				_, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
				assert.NoError(t, err, "Valid png qr code")

				user, err := us.GetByEmail(tt.email)
				assert.NotEmpty(t, user.OtpQrCode)
				assert.NotEmpty(t, user.OtpSecret)
				assert.False(t, user.TwoFaOTPEnabled)
			} else {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))
			}
		})
	}
}

func TestHandler_VerifyOTPCode(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewInMemory()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}
	db.Clear(database)

	us := db.NewUserStore(database)

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.Data.DomainName,
		AccountName: "paul",
	})
	var qr bytes.Buffer
	img, err := key.Image(200, 200)
	png.Encode(&qr, img)

	us.Create(&models.User{
		Name:      "bob",
		Email:     "bob@gmail.com",
		OtpSecret: key.Secret(),
		OtpQrCode: qr.Bytes(),
	})

	pastCode, _ := totp.GenerateCode(key.Secret(), time.UnixMilli(0))
	currentCode, _ := totp.GenerateCode(key.Secret(), time.Now())

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		otp         string
		loginToken  string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", otp: currentCode, loginToken: "", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Correct code"},
		{tName: "Wrong email", email: "bobo@gmail.com", otp: currentCode, loginToken: "", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong otp code", email: "bob@gmail.com", otp: pastCode, loginToken: "", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "With login token", email: "bob@gmail.com", otp: currentCode, loginToken: "asdfasdfsadfasdf", wantCode: http.StatusNotImplemented, wantSuccess: false, wantMessage: "Not yet implemented"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "otp_code": "%s", "login_token": "%s"}`, tt.email, tt.otp, tt.loginToken)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.VerifyOTPCode(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))
		})
	}
}

func TestHandler_Login(t *testing.T) {
	config.Data.Debug = true
	r := router.New()

	database, err := db.NewInMemory()
	if err != nil {
		t.Fatalf("Couldn't create in memory database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}
	db.Clear(database)

	us := db.NewUserStore(database)

	password, err := bcrypt.GenerateFromPassword([]byte("password"), config.Data.BcryptCost)
	us.Create(&models.User{
		Name:            "bob",
		Email:           "bob@gmail.com",
		PasswordHash:    password,
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
	})

	us.Create(&models.User{
		Name:           "paul",
		Email:          "paul@gmail.com",
		PasswordHash:   password,
		EmailConfirmed: true,
	})

	us.Create(&models.User{
		Name:            "peter",
		Email:           "peter@gmail.com",
		PasswordHash:    password,
		TwoFaOTPEnabled: true,
	})

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", password: "password", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully signed in"},
		{tName: "Wrong email", email: "bobo@gmail.com", password: "password", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong password", email: "bob@gmail.com", password: "drowssap", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "2FA not enabled", email: "paul@gmail.com", password: "password", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "2FA is not enabled"},
		{tName: "Email not confirmed", email: "peter@gmail.com", password: "password", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Email is not confirmed"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, tt.email, tt.password)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.Login(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), `"login_token":`)

				user, _ := us.GetByEmail(tt.email)
				tokens, _ := us.GetLoginTokens(user)
				assert.Equal(t, 1, len(tokens), "A login token was stored in the database")
			}
		})
	}
}
