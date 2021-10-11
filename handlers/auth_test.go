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
	"github.com/google/uuid"
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
			ExpirationTime: time.Now().Unix() + config.Data.EmailCodeLifetime,
		},
	})

	us.Create(&models.User{
		Name:  "paul",
		Email: "paul@gmail.com",
		EmailCode: models.EmailCode{
			Code:           "123456",
			ExpirationTime: time.Now().Unix() + config.Data.EmailCodeLifetime,
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

func TestHandler_GetOTPQRCode(t *testing.T) {
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

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), config.Data.BcryptCost)
	user := &models.User{
		PasswordHash: hash,
		OtpQrCode:    []byte("png_qr_code"),
	}
	us.Create(user)

	handler := New(us)

	tests := []struct {
		tName       string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", password: "123456", wantCode: http.StatusOK, wantSuccess: true},
		{tName: "Wrong password", password: "654321", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"password": "%s"}`, tt.password)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			c.Set("userId", user.Id)

			err := handler.GetOTPQRCode(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			if !tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))
			} else {
				assert.Equal(t, "png_qr_code", rec.Body.String())
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

	pastCode, _ := totp.GenerateCode(key.Secret(), time.Unix(0, 0))
	currentCode, _ := totp.GenerateCode(key.Secret(), time.Now())

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		otp         string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", otp: currentCode, wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully aquired two factor token"},
		{tName: "Wrong email", email: "bobo@gmail.com", otp: currentCode, wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong otp code", email: "bob@gmail.com", otp: pastCode, wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "code": "%s"}`, tt.email, tt.otp)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.VerifyOTPCode(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), `"token":`)

				user, _ := us.GetByEmail(tt.email)
				tokens, _ := us.GetTwoFATokens(user)
				assert.Equal(t, 1, len(tokens), "A two factor token was stored in the database")
			}
		})
	}
}

func TestHandler_PasswordAuth(t *testing.T) {
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

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", password: "password", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully aquired password token"},
		{tName: "Wrong email", email: "bobo@gmail.com", password: "password", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong password", email: "bob@gmail.com", password: "drowssap", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, tt.email, tt.password)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.PasswordAuth(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), `"token":`)

				user, _ := us.GetByEmail(tt.email)
				tokens, _ := us.GetPasswordTokens(user)
				assert.Equal(t, 1, len(tokens), "A password token was stored in the database")
			}
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

	us.Create(&models.User{
		Name:            "bob",
		Email:           "bob@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
		PasswordTokens:  []models.PasswordToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
		TwoFATokens:     []models.TwoFAToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
	})

	us.Create(&models.User{
		Name:            "tim",
		Email:           "tim@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
		PasswordTokens:  []models.PasswordToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
		TwoFATokens:     []models.TwoFAToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
	})

	us.Create(&models.User{
		Name:            "paul",
		Email:           "paul@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
		PasswordTokens:  []models.PasswordToken{{Code: "1234567890", ExpirationTime: 0}},
		TwoFATokens:     []models.TwoFAToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
	})

	us.Create(&models.User{
		Name:            "peter",
		Email:           "peter@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
		PasswordTokens:  []models.PasswordToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
		TwoFATokens:     []models.TwoFAToken{{Code: "1234567890", ExpirationTime: 0}},
	})

	us.Create(&models.User{
		Name:            "hans",
		Email:           "hans@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
		PasswordTokens:  []models.PasswordToken{{Code: "1234567890", ExpirationTime: 0}},
		TwoFATokens:     []models.TwoFAToken{{Code: "1234567890", ExpirationTime: 0}},
	})

	handler := New(us)

	tests := []struct {
		tName          string
		email          string
		passwordToken  string
		twoFactorToken string
		wantCode       int
		wantSuccess    bool
		wantMessage    string
	}{
		{tName: "Success", email: "bob@gmail.com", passwordToken: "1234567890", twoFactorToken: "1234567890", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully signed in"},
		{tName: "Wrong email", email: "tom@gmail.com", passwordToken: "1234567890", twoFactorToken: "1234567890", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong password token", email: "tim@gmail.com", passwordToken: "0987654321", twoFactorToken: "1234567890", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong two factor token", email: "tim@gmail.com", passwordToken: "1234567890", twoFactorToken: "0987654321", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Expired password token", email: "paul@gmail.com", passwordToken: "1234567890", twoFactorToken: "1234567890", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Expired two factor token", email: "peter@gmail.com", passwordToken: "1234567890", twoFactorToken: "1234567890", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Both tokens expired", email: "peter@gmail.com", passwordToken: "1234567890", twoFactorToken: "1234567890", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "password_token": "%s", "two_fa_token": "%s"}`, tt.email, tt.passwordToken, tt.twoFactorToken)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.Login(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			user, _ := us.GetByEmail(tt.email)
			if user != nil {
				if tt.wantSuccess {
					cookies := rec.Result().Cookies()
					assert.Equal(t, 3, len(cookies), "Three auth cookies were returned")
					for _, cookie := range cookies {
						assert.True(t, cookie.Secure)
						assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
					}

					refreshTokens, _ := us.GetRefreshTokens(user)
					assert.Equal(t, 1, len(refreshTokens), "A refresh token was stored in the database")
				}

				if tt.passwordToken == "1234567890" && tt.twoFactorToken == "1234567890" {
					pTokens, _ := us.GetPasswordTokens(user)
					if len(pTokens) == 1 {
						assert.True(t, pTokens[0].ExpirationTime > time.Now().Unix())
					}
					tFATokens, _ := us.GetTwoFATokens(user)
					if len(tFATokens) == 1 {
						assert.True(t, tFATokens[0].ExpirationTime > time.Now().Unix())
					}
				}
			}
		})
	}
}

func TestHandler_GetRecoveryCodes(t *testing.T) {
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

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), config.Data.BcryptCost)
	user := &models.User{
		PasswordHash: hash,
		RecoveryCodes: []models.RecoveryCode{
			{Code: "öareoghöaorwenhgöareohgoaöwrhgaeorgha"},
			{Code: "askfjaösdhfgoöasdhfoöasdhföasdhfökjas"},
			{Code: "aslkfjöasdjfjasbdviusadhföalsjdhföasd"},
			{Code: "öasdfhsuighösafnöasjdföashdgoaösdfkjd"},
			{Code: "öalskfsaoghskfnöosauhgpisejfäsgjösadd"},
		},
	}
	us.Create(user)

	handler := New(us)

	tests := []struct {
		tName       string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", password: "123456", wantCode: http.StatusOK, wantSuccess: true},
		{tName: "Wrong password", password: "654321", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"password": "%s"}`, tt.password)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			c.Set("userId", user.Id)

			err := handler.GetRecoveryCodes(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				codes := user.RecoveryCodes
				for _, c := range codes {
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"%s"`, c.Code))
				}
			}
		})
	}
}

func TestHandler_VerifyRecoveryCode(t *testing.T) {
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
		RecoveryCodes: []models.RecoveryCode{
			{Code: "1234567890"},
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
		{tName: "Success", email: "bob@gmail.com", code: "1234567890", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully aquired two factor token"},
		{tName: "Wrong email", email: "bobo@gmail.com", code: "1234567890", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong recovery code", email: "bob@gmail.com", code: "0987654321", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "code": "%s"}`, tt.email, tt.code)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.VerifyRecoveryCode(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), `"token":`)

				user, _ := us.GetByEmail(tt.email)

				tokens, _ := us.GetTwoFATokens(user)
				assert.Equal(t, 1, len(tokens), "A two factor token was stored in the database")

				rCodes, _ := us.GetRecoveryCodes(user)
				assert.Empty(t, rCodes, "The recovery code was deleted from the database")
			}
		})
	}
}

func TestHandler_NewRecoveryCodes(t *testing.T) {
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

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), config.Data.BcryptCost)
	user := &models.User{
		PasswordHash: hash,
		RecoveryCodes: []models.RecoveryCode{
			{Code: "öareoghöaorwenhgöareohgoaöwrhgaeorgha"},
			{Code: "askfjaösdhfgoöasdhfoöasdhföasdhfökjas"},
			{Code: "aslkfjöasdjfjasbdviusadhföalsjdhföasd"},
			{Code: "öasdfhsuighösafnöasjdföashdgoaösdfkjd"},
			{Code: "lalskfsaoghskfnöosauhgpisejfäsgjösadd"},
			{Code: "zalskfsaoghskfnöosauhgpisejfäsgjösadd"},
			{Code: "oalskfsaoghskfnöosauhgpisejfäsgjösadd"},
			{Code: "aalskfsaoghskfnöosauhgpisejfäsgjösadd"},
			{Code: "üalskfsaoghskfnöosauhgpisejfäsgjösadd"},
			{Code: "jalskfsaoghskfnöosauhgpisejfäsgjösadd"},
		},
	}
	us.Create(user)

	handler := New(us)

	tests := []struct {
		tName       string
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", password: "123456", wantCode: http.StatusOK, wantSuccess: true},
		{tName: "Wrong password", password: "654321", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"password": "%s"}`, tt.password)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			c.Set("userId", user.Id)

			err := handler.NewRecoveryCodes(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				prevCodes := user.RecoveryCodes
				codes, _ := us.GetRecoveryCodes(user)

				for i, c := range codes {
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"%s"`, c.Code))
					assert.NotContains(t, rec.Body.String(), fmt.Sprintf(`"%s"`, prevCodes[i].Code))

					for _, c2 := range prevCodes {
						assert.NotEqual(t, c.Code, c2.Code)
					}
				}
			}
		})
	}
}

func TestHandler_NewOTP(t *testing.T) {
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

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), config.Data.BcryptCost)
	user1 := &models.User{
		Name:            "bob",
		Email:           "bob@gmail.com",
		PasswordHash:    hash,
		OtpQrCode:       []byte("png_qr_code"),
		TwoFaOTPEnabled: true,
	}
	us.Create(user1)

	user2 := &models.User{
		Name:            "peter",
		Email:           "peter@gmail.com",
		PasswordHash:    hash,
		TwoFaOTPEnabled: false,
		OtpQrCode:       []byte("png_qr_code"),
	}
	us.Create(user2)

	handler := New(us)

	tests := []struct {
		tName       string
		user        *models.User
		password    string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", user: user1, password: "123456", wantCode: http.StatusOK, wantSuccess: true},
		{tName: "Wrong password", user: user1, password: "654321", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "OTP not enabled", user: user2, password: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Please enable otp first"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"password": "%s"}`, tt.password)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			c.Set("userId", tt.user.Id)

			err := handler.NewOTP(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			if !tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))
			} else {
				assert.Equal(t, "image/png", rec.HeaderMap.Get("Content-Type"))
				assert.NotEmpty(t, rec.Body)
				assert.NotEqual(t, "png_qr_code", rec.Body.String())
			}
		})
	}
}

func TestHandler_Logout(t *testing.T) {
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

	codeStr1 := "sadhfasdhfasdhjfsaliudlhfaskjfdhlasid"
	codeStr2 := "asödfjasiefjsöalkejföosiaefjölaskejfs"
	code1, _ := bcrypt.GenerateFromPassword([]byte(codeStr1), config.Data.BcryptCost)
	code2, _ := bcrypt.GenerateFromPassword([]byte(codeStr2), config.Data.BcryptCost)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		RefreshTokens: []models.RefreshToken{
			{Code: code1},
			{Code: code2},
		},
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "peter",
		Email: "peter@gmail.com",
		RefreshTokens: []models.RefreshToken{
			{Code: code1},
			{Code: code2},
		},
	}
	us.Create(user2)

	handler := New(us)

	tests := []struct {
		tName       string
		user        *models.User
		all         bool
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Not all", user: user1, all: false, wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully signed out"},
		{tName: "All", user: user1, all: true, wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully signed out"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/?all=%t", tt.all), nil)

			rTokens, _ := us.GetRefreshTokens(tt.user)
			req.AddCookie(&http.Cookie{
				Name:  "Refresh-Token",
				Value: rTokens[0].Id.String() + codeStr1,
			})
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			c.Set("userId", tt.user.Id)

			err := handler.Logout(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			refreshTokens, _ := us.GetRefreshTokens(tt.user)
			if tt.all {
				assert.Empty(t, refreshTokens)
			} else {
				assert.Equal(t, 1, len(refreshTokens))
			}
		})
	}
}

func TestHandler_ChangePassword(t *testing.T) {
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

	password := "123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), config.Data.BcryptCost)
	user1 := &models.User{
		Name:         "bob",
		Email:        "bob@gmail.com",
		PasswordHash: hash,
		TwoFATokens:  []models.TwoFAToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
	}
	us.Create(user1)

	user2 := &models.User{
		Name:         "bob2",
		Email:        "bob2@gmail.com",
		PasswordHash: hash,
		TwoFATokens:  []models.TwoFAToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
	}
	us.Create(user2)

	user3 := &models.User{
		Name:         "bob3",
		Email:        "bob3@gmail.com",
		PasswordHash: hash,
		TwoFATokens:  []models.TwoFAToken{{Code: "1234567890", ExpirationTime: 0}},
	}
	us.Create(user3)

	handler := New(us)

	tests := []struct {
		tName       string
		user        *models.User
		twoFAToken  string
		oldPassword string
		newPassword string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", user: user1, twoFAToken: "1234567890", oldPassword: "123456", newPassword: "abcdef", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully changed password"},
		{tName: "Wrong password", user: user2, twoFAToken: "1234567890", oldPassword: "654321", newPassword: "abcdef", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "New password too short", user: user2, twoFAToken: "1234567890", oldPassword: "123456", newPassword: "abcde", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "New password too short (min 6)"},
		{tName: "New password too long", user: user2, twoFAToken: "1234567890", oldPassword: "123456", newPassword: strings.Repeat("a", 70), wantCode: http.StatusOK, wantSuccess: false, wantMessage: "New password too long (max 64)"},
		{tName: "Wrong two factor token", user: user2, twoFAToken: "0987654321", oldPassword: "123456", newPassword: "abcdef", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Expired two factor token", user: user3, twoFAToken: "1234567890", oldPassword: "123456", newPassword: "abcdef", wantCode: http.StatusForbidden, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"old_password": "%s", "new_password": "%s", "two_fa_token": "%s"}`, tt.oldPassword, tt.newPassword, tt.twoFAToken)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			c.Set("userId", tt.user.Id)

			err := handler.ChangePassword(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			user, _ := us.GetById(tt.user.Id)
			if tt.wantSuccess {
				assert.Error(t, bcrypt.CompareHashAndPassword(user.PasswordHash, []byte("123456")))
				assert.NoError(t, bcrypt.CompareHashAndPassword(user.PasswordHash, []byte("abcdef")))
			} else {
				assert.NoError(t, bcrypt.CompareHashAndPassword(user.PasswordHash, []byte("123456")))
				assert.Error(t, bcrypt.CompareHashAndPassword(user.PasswordHash, []byte("abcdef")))
			}

			if tt.twoFAToken == "1234567890" && tt.oldPassword == "123456" && tt.newPassword == "abcdef" {
				tokens, _ := us.GetTwoFATokens(user)
				assert.Empty(t, tokens)
			}
		})
	}
}

func TestHandler_ForgotPassword(t *testing.T) {
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
		TwoFATokens: []models.TwoFAToken{
			{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime},
			{Code: "12345678901", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime},
		},
	})

	us.Create(&models.User{
		Name:        "peter",
		Email:       "peter@gmail.com",
		TwoFATokens: []models.TwoFAToken{{Code: "1234567890", ExpirationTime: 0}},
	})

	us.Create(&models.User{
		Name:        "paul",
		Email:       "paul@gmail.com",
		TwoFATokens: []models.TwoFAToken{{Code: "1234567890", ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime}},
	})

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		twoFAToken  string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", twoFAToken: "1234567890", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "An email with a reset password link has been sent to the specified address"},
		{tName: "Expired two factor token", email: "peter@gmail.com", twoFAToken: "1234567890", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong two factor token", email: "paul@gmail.com", twoFAToken: "0987654321", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Non existing user", email: "hans@gmail.com", twoFAToken: "0987654321", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "two_fa_token": "%s"})`, tt.email, tt.twoFAToken)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.ForgotPassword(c)

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

					jsonBody := fmt.Sprintf(`{"email": "%s", "two_fa_token": "%s"})`, tt.email, "12345678901")
					req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					c := r.NewContext(req, rec)

					err = handler.ForgotPassword(c)

					assert.NoError(t, err)
					assert.Equal(t, http.StatusTooManyRequests, rec.Code)
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, false))
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, "Please wait at least 2 minutes between forgot password email requests"))
				}
			}
		})
	}
}

func TestHandler_ResetPassword(t *testing.T) {
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

	user1 := &models.User{
		Name:      "bob",
		Email:     "bob@gmail.com",
		EmailCode: models.EmailCode{Code: "abcdefg", ExpirationTime: time.Now().Unix() + config.Data.EmailCodeLifetime},
	}
	us.Create(user1)

	user2 := &models.User{
		Name:      "bob2",
		Email:     "bob2@gmail.com",
		EmailCode: models.EmailCode{Code: "abcdefg", ExpirationTime: time.Now().Unix() + config.Data.EmailCodeLifetime},
	}
	us.Create(user2)

	user3 := &models.User{
		Name:      "bob3",
		Email:     "bob3@gmail.com",
		EmailCode: models.EmailCode{Code: "abcdefg", ExpirationTime: 0},
	}
	us.Create(user3)

	handler := New(us)

	tests := []struct {
		tName       string
		email       string
		token       string
		newPassword string
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", email: "bob@gmail.com", token: "abcdefg", newPassword: "123456", wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully changed password"},
		{tName: "Expired token", email: "bob3@gmail.com", token: "abcdefg", newPassword: "123456", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Non existing user", email: "bob4@gmail.com", token: "abcdefg", newPassword: "123456", wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Password too short", email: "bob2@gmail.com", token: "abcdefg", newPassword: "12345", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "New password too short (min 6)"},
		{tName: "Password too long", email: "bob2@gmail.com", token: "abcdefg", newPassword: strings.Repeat("a", 70), wantCode: http.StatusOK, wantSuccess: false, wantMessage: "New password too long (max 64)"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"email": "%s", "new_password": "%s", "token": "%s"}`, tt.email, tt.newPassword, tt.token)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			err := handler.ResetPassword(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			user, _ := us.GetByEmail(tt.email)
			if tt.wantSuccess {
				assert.NoError(t, bcrypt.CompareHashAndPassword(user.PasswordHash, []byte("123456")))
			}

			if user != nil && tt.token == "abcdefg" && tt.newPassword == "123456" {
				code, err := us.GetEmailCode(user)
				assert.NoError(t, err)
				assert.Nil(t, code)
			}
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
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

	codeStr1 := "sadhfasdhfasdhjfsaliudlhfaskjfdhlasid"
	codeStr2 := "asudfjasiefjsualkejfuosiaefjulaskejfs"
	code1, _ := bcrypt.GenerateFromPassword([]byte(codeStr1), config.Data.BcryptCost)
	code2, _ := bcrypt.GenerateFromPassword([]byte(codeStr2), config.Data.BcryptCost)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		RefreshTokens: []models.RefreshToken{
			{Code: code1},
			{Code: code2, Used: true},
		},
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "peter",
		Email: "peter@gmail.com",
		RefreshTokens: []models.RefreshToken{
			{Code: code1},
			{Code: code2, Used: true},
		},
	}
	us.Create(user2)

	user3 := &models.User{
		Name:  "paul",
		Email: "paul@gmail.com",
		RefreshTokens: []models.RefreshToken{
			{Code: code1},
			{Code: code2, Used: true},
		},
	}
	us.Create(user1)

	handler := New(us)

	tests := []struct {
		tName            string
		userId           uuid.UUID
		refreshCode      string
		refreshCodeIndex int
		wantCode         int
		wantSuccess      bool
		wantMessage      string
	}{
		{tName: "Success", userId: user1.Id, refreshCode: codeStr1, refreshCodeIndex: 0, wantCode: http.StatusOK, wantSuccess: true, wantMessage: "Successfully refreshed tokens"},
		{tName: "Used token", userId: user3.Id, refreshCode: codeStr2, refreshCodeIndex: 1, wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Non existing user", userId: uuid.New(), refreshCode: codeStr1, refreshCodeIndex: 0, wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
		{tName: "Wrong token code", userId: user2.Id, refreshCode: "asudfjasiefjsualkejfuosiaefjulaskejfs", refreshCodeIndex: 0, wantCode: http.StatusUnauthorized, wantSuccess: false, wantMessage: "Invalid credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"user_id": "%s"}`, tt.userId.String())
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			user, _ := us.GetById(tt.userId)
			if user != nil {
				rTokens, _ := us.GetRefreshTokens(user)
				req.AddCookie(&http.Cookie{
					Name:  "Refresh-Token",
					Value: rTokens[tt.refreshCodeIndex].Id.String() + tt.refreshCode,
				})

				rec := httptest.NewRecorder()
				c := r.NewContext(req, rec)

				err := handler.Refresh(c)

				assert.NoError(t, err)
				assert.Equal(t, tt.wantCode, rec.Code)
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

				if tt.wantSuccess {
					cookies := rec.Result().Cookies()
					assert.Equal(t, 3, len(cookies), "Three new auth cookies were returned")
					for _, cookie := range cookies {
						assert.True(t, cookie.Secure)
						assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
					}

					refreshTokens, _ := us.GetRefreshTokens(user)
					assert.Equal(t, 3, len(refreshTokens), "A new refresh token has been created")

					notUsed := 0
					for _, r := range refreshTokens {
						if !r.Used {
							notUsed++
						}
					}
					assert.Equal(t, 1, notUsed, "The old refresh token has been marked as used")
				}

				if tt.refreshCodeIndex == 1 {
					refreshTokens, _ := us.GetRefreshTokens(user)
					assert.Empty(t, refreshTokens, "All refresh tokens were deleted")
				}
			}
		})
	}
}
