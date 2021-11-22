package handlers

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Bananenpro/hbank-api/bindings"
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// /v1/auth/register (POST)
func (h *Handler) Register(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.Register
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if !services.VerifyCaptcha(body.CaptchaToken) {
		return c.JSON(http.StatusOK, responses.New(false, "Invalid captcha token", lang))
	}

	body.Name = strings.ToLower(body.Name)
	body.Email = strings.ToLower(body.Email)

	if !services.IsValidEmail(body.Email) {
		return c.JSON(http.StatusOK, responses.New(false, "Invalid email", lang))
	}

	if len(body.Name) > config.Data.MaxNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Name too long", lang))
	}

	if utf8.RuneCountInString(body.Name) < config.Data.MinNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Name too short", lang))
	}

	if len(body.Password) > config.Data.MaxPasswordLength {
		return c.JSON(http.StatusOK, responses.New(false, "Password too long", lang))
	}

	if utf8.RuneCountInString(body.Password) < config.Data.MinPasswordLength {
		return c.JSON(http.StatusOK, responses.New(false, "Password too short", lang))
	}

	if u, _ := h.userStore.GetByEmail(body.Email); u != nil {
		return c.JSON(http.StatusOK, responses.New(false, "The user with this email does already exist", lang))
	}

	user := &models.User{
		Name:             body.Name,
		Email:            body.Email,
		ProfilePictureId: uuid.New(),
	}

	var err error
	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(body.Password), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	err = h.userStore.Create(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusCreated, responses.NewAuthUser(user))
}

// /v1/auth/confirmEmail/:email (GET)
func (h *Handler) SendConfirmEmail(c echo.Context) error {
	lang := c.Get("lang").(string)
	email := c.Param("email")
	if !services.IsValidEmail(email) {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Missing or invalid email parameter", lang))
	}

	lastSend, err := h.userStore.GetConfirmEmailLastSent(email)
	if lastSend+config.Data.SendEmailTimeout > time.Now().Unix() {
		return c.JSON(http.StatusTooManyRequests, responses.Base{
			Success: false,
			Message: fmt.Sprintf(services.Tr("Please wait at least %d minutes between confirm email requests", lang), config.Data.SendEmailTimeout/60),
		})
	}
	h.userStore.SetConfirmEmailLastSent(email, time.Now().Unix())

	user, err := h.userStore.GetByEmail(email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user != nil {
		emailCode, err := h.userStore.GetConfirmEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if !user.EmailConfirmed {
			h.userStore.DeleteConfirmEmailCode(emailCode)
			code := services.GenerateRandomString(6)
			user.ConfirmEmailCode = models.ConfirmEmailCode{
				CodeHash: services.HashToken(code),
			}
			err = h.userStore.Update(user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}

			if config.Data.EmailEnabled {
				type templateData struct {
					Name      string
					Code      string
					DeleteUrl string
				}
				body, err := services.ParseEmailTemplate("confirmEmail", c.Get("lang").(string), templateData{
					Name:      user.Name,
					Code:      code,
					DeleteUrl: fmt.Sprintf("https://%s/account/delete?code=%s", config.Data.DomainName, code),
				})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
				}
				go services.SendEmail([]string{user.Email}, services.Tr("H-Bank Confirm Email", lang), body)
			}
		}
	}

	return c.JSON(http.StatusOK, responses.New(true, "If the email address is linked to a user whose email has not yet been confirmed, a code has been sent to the specified address", lang))
}

// /v1/auth/confirmEmail (POST)
func (h *Handler) VerifyConfirmEmailCode(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.ConfirmEmail
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if user != nil {
		confirmEmailCode, err := h.userStore.GetConfirmEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if confirmEmailCode != nil {
			if subtle.ConstantTimeCompare(confirmEmailCode.CodeHash, services.HashToken(body.Code)) == 1 {
				user.EmailConfirmed = true
				err = h.userStore.Update(user)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
				}

				h.userStore.DeleteConfirmEmailCode(confirmEmailCode)
				return c.JSON(http.StatusOK, responses.New(true, "Successfully confirmed email address", lang))
			}
		}
	}

	return c.JSON(http.StatusOK, responses.New(false, "Email was not confirmed", lang))
}

// /v1/auth/twoFactor/otp/activate (POST)
func (h *Handler) Activate2FAOTP(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.EmailPassword
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if !user.TwoFaOTPEnabled {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      config.Data.DomainName,
			AccountName: user.Email,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		img, err := key.Image(200, 200)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		var qr bytes.Buffer

		png.Encode(&qr, img)

		secret := key.Secret()

		user.OtpSecret = secret
		user.OtpQrCode = qr.Bytes()

		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		return c.JSON(http.StatusOK, responses.New(true, "Successfully activated TwoFaOTP", lang))
	}

	return c.JSON(http.StatusOK, responses.New(false, "TwoFaOTP is already activated", lang))
}

// /v1/auth/twoFactor/otp/qr (POST)
func (h *Handler) GetOTPQRCode(c echo.Context) error {
	lang := c.Get("lang").(string)

	var body bindings.EmailPassword
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	if !user.TwoFaOTPEnabled {
		return c.Blob(http.StatusOK, "image/png", user.OtpQrCode)
	} else {
		return c.JSON(http.StatusOK, responses.New(false, "TwoFaOTP is already activated", lang))
	}
}

// /v1/auth/twoFactor/otp/key (POST)
func (h *Handler) GetOTPKey(c echo.Context) error {
	lang := c.Get("lang").(string)

	var body bindings.EmailPassword
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	if !user.TwoFaOTPEnabled {
		return c.JSON(http.StatusOK, responses.Token{
			Base: responses.Base{
				Success: true,
			},
			Token: user.OtpSecret,
		})
	} else {
		return c.JSON(http.StatusOK, responses.New(false, "TwoFaOTP is already activated", lang))
	}
}

// /v1/auth/twoFactor/otp/verify (POST)
func (h *Handler) VerifyOTPCode(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.VerifyCode
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if totp.Validate(body.Code, user.OtpSecret) {
		user.TwoFaOTPEnabled = true
		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		code := ""
		exists := true

		for exists {
			code = services.GenerateRandomString(64)
			t, err := h.userStore.GetTwoFATokenByCode(user, code)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			exists = t != nil
		}

		user.TwoFATokens = append(user.TwoFATokens, models.TwoFAToken{
			CodeHash:       services.HashToken(code),
			ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime,
		})
		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		return c.JSON(http.StatusOK, responses.Token{
			Base: responses.Base{
				Success: true,
				Message: services.Tr("Successfully aquired two factor token", lang),
			},
			Token: code,
		})
	}

	return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
}

// /v1/auth/twoFactor/otp/new
func (h *Handler) NewOTP(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.Password
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	if user.TwoFaOTPEnabled {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      config.Data.DomainName,
			AccountName: user.Email,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		img, err := key.Image(200, 200)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		var qr bytes.Buffer

		png.Encode(&qr, img)

		secret := key.Secret()

		user.OtpSecret = secret
		user.OtpQrCode = qr.Bytes()

		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		return c.JSON(http.StatusOK, responses.New(true, "Successfully created new otp", lang))
	}

	return c.JSON(http.StatusOK, responses.New(false, "Please enable otp first", lang))
}

// /v1/auth/passwordAuth (POST)
func (h *Handler) PasswordAuth(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.PasswordAuth
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	code := ""
	exists := true

	for exists {
		code = services.GenerateRandomString(64)
		t, err := h.userStore.GetPasswordTokenByCode(user, code)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		exists = t != nil
	}

	user.PasswordTokens = append(user.PasswordTokens, models.PasswordToken{
		CodeHash:       services.HashToken(code),
		ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime,
	})
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.Token{
		Base: responses.Base{
			Success: true,
			Message: services.Tr("Successfully aquired password token", lang),
		},
		Token: code,
	})
}

// /v1/auth/login (POST)
func (h *Handler) Login(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.Login
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	passwordToken, err := h.userStore.GetPasswordTokenByCode(user, body.PasswordToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if passwordToken == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if twoFAToken == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if passwordToken.ExpirationTime < time.Now().Unix() {
		h.userStore.DeletePasswordToken(passwordToken)
	}
	if twoFAToken.ExpirationTime < time.Now().Unix() {
		h.userStore.DeleteTwoFAToken(twoFAToken)
	}
	if passwordToken.ExpirationTime < time.Now().Unix() || twoFAToken.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	h.userStore.DeletePasswordToken(passwordToken)
	h.userStore.DeleteTwoFAToken(twoFAToken)

	if !user.EmailConfirmed {
		return c.JSON(http.StatusOK, responses.New(false, "Email is not confirmed", lang))
	}

	if !user.TwoFaOTPEnabled {
		return c.JSON(http.StatusOK, responses.New(false, "Two factor authentication is not enabled", lang))
	}

	code := services.GenerateRandomString(64)
	hash, err := bcrypt.GenerateFromPassword([]byte(code), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	refreshToken := &models.RefreshToken{
		CodeHash:       hash,
		ExpirationTime: time.Now().Unix() + config.Data.RefreshTokenLifetime,
	}
	err = h.userStore.AddRefreshToken(user, refreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	authToken, authTokenSignature, err := services.NewAuthToken(user)
	if err != nil {
		h.userStore.DeleteRefreshToken(refreshToken)
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	c.SetCookie(&http.Cookie{
		Name:     "Refresh-Token",
		Value:    refreshToken.Id.String() + code,
		MaxAge:   int(config.Data.RefreshTokenLifetime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth",
	})

	c.SetCookie(&http.Cookie{
		Name:     "Auth-Token",
		Value:    authToken,
		MaxAge:   int(config.Data.AuthTokenLifetime),
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1",
	})

	c.SetCookie(&http.Cookie{
		Name:     "Auth-Token-Signature",
		Value:    authTokenSignature,
		MaxAge:   int(config.Data.AuthTokenLifetime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1",
	})

	return c.JSON(http.StatusOK, responses.NewAuthUser(user))
}

// /v1/auth/refresh (POST)
func (h *Handler) Refresh(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.Refresh
	err := c.Bind(&body)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	userId, err := uuid.Parse(body.UserId)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	refreshCookie, err := c.Cookie("Refresh-Token")
	if err != nil || len(refreshCookie.Value) <= 36 {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	refreshTokenId, err := uuid.Parse(refreshCookie.Value[:36])
	if err != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	refreshToken, err := h.userStore.GetRefreshToken(user, refreshTokenId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if refreshToken == nil || bcrypt.CompareHashAndPassword(refreshToken.CodeHash, []byte(refreshCookie.Value[36:])) != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if refreshToken.ExpirationTime < time.Now().Unix() {
		err = h.userStore.DeleteRefreshToken(refreshToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	if refreshToken.Used {
		err = h.userStore.DeleteRefreshTokens(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	newRefreshToken, code, err := h.userStore.RotateRefreshToken(user, refreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	authToken, authTokenSignature, err := services.NewAuthToken(user)
	if err != nil {
		h.userStore.DeleteRefreshToken(newRefreshToken)
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	c.SetCookie(&http.Cookie{
		Name:     "Refresh-Token",
		Value:    newRefreshToken.Id.String() + code,
		MaxAge:   int(config.Data.RefreshTokenLifetime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth",
	})

	c.SetCookie(&http.Cookie{
		Name:     "Auth-Token",
		Value:    authToken,
		MaxAge:   int(config.Data.AuthTokenLifetime),
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1",
	})

	c.SetCookie(&http.Cookie{
		Name:     "Auth-Token-Signature",
		Value:    authTokenSignature,
		MaxAge:   int(config.Data.AuthTokenLifetime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1",
	})

	return c.JSON(http.StatusOK, responses.New(true, "Successfully refreshed tokens", lang))
}

// /v1/auth/logout?all=bool (POST)
func (h *Handler) Logout(c echo.Context) error {
	lang := c.Get("lang").(string)
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	if services.StrToBool(c.QueryParams().Get("all")) {
		err = h.userStore.DeleteRefreshTokens(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
	} else {
		refreshCookie, err := c.Cookie("Refresh-Token")
		if err != nil || len(refreshCookie.Value) <= 36 {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		tokenId, err := uuid.Parse(refreshCookie.Value[:36])
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		refreshToken, err := h.userStore.GetRefreshToken(user, tokenId)
		if err != nil || refreshToken == nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		err = h.userStore.DeleteRefreshToken(refreshToken)
		if err != nil || refreshToken == nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully signed out", lang))
}

// /v1/auth/twoFactor/recovery/verify (POST)
func (h *Handler) VerifyRecoveryCode(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.VerifyCode
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	code, err := h.userStore.GetRecoveryCodeByCode(user, body.Code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if code == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	token := ""
	exists := true

	for exists {
		token = services.GenerateRandomString(64)
		t, err := h.userStore.GetTwoFATokenByCode(user, token)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		exists = t != nil
	}

	err = h.userStore.DeleteRecoveryCode(code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	user.TwoFATokens = append(user.TwoFATokens, models.TwoFAToken{
		CodeHash:       services.HashToken(token),
		ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime,
	})
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.Token{
		Base: responses.Base{
			Success: true,
			Message: services.Tr("Successfully aquired two factor token", lang),
		},
		Token: token,
	})
}

// /v1/auth/twoFactor/recovery/new (POST)
func (h *Handler) NewRecoveryCodes(c echo.Context) error {
	lang := c.Get("lang").(string)
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.Password
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	codes, err := h.userStore.NewRecoveryCodes(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.RecoveryCodes{
		Base: responses.Base{
			Success: true,
		},
		Codes: codes,
	})
}

// /v1/auth/changePassword (POST)
func (h *Handler) ChangePassword(c echo.Context) error {
	lang := c.Get("lang").(string)
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.ChangePassword
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if len(body.NewPassword) > config.Data.MaxPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf(services.Tr("New password too long (max %d)", lang), config.Data.MaxPasswordLength),
		})
	}

	if utf8.RuneCountInString(body.NewPassword) < config.Data.MinPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf(services.Tr("New password too short (min %d)", lang), config.Data.MinPasswordLength),
		})
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.OldPassword)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(body.NewPassword), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully changed password", lang))
}

// /v1/auth/forgotPassword (POST)
func (h *Handler) ForgotPassword(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.ForgotPassword
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if !services.IsValidEmail(body.Email) {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid email", lang))
	}

	if services.VerifyCaptcha(body.CaptchaToken) {
		user, err := h.userStore.GetByEmail(body.Email)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if user == nil {
			return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
		}
		twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if twoFAToken == nil {
			return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
		}
		h.userStore.DeleteTwoFAToken(twoFAToken)
		if twoFAToken.ExpirationTime < time.Now().Unix() {
			return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
		}

		lastSend, err := h.userStore.GetForgotPasswordEmailLastSent(body.Email)
		if lastSend+config.Data.SendEmailTimeout > time.Now().Unix() {
			return c.JSON(http.StatusTooManyRequests, responses.Base{
				Success: false,
				Message: fmt.Sprintf(services.Tr("Please wait at least %d minutes between forgot password email requests", lang), config.Data.SendEmailTimeout/60),
			})
		}
		h.userStore.SetForgotPasswordEmailLastSent(body.Email, time.Now().Unix())

		emailCode, err := h.userStore.GetResetPasswordCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		h.userStore.DeleteResetPasswordCode(emailCode)
		code := services.GenerateRandomString(64)
		user.ResetPasswordCode = models.ResetPasswordCode{
			CodeHash:       services.HashToken(code),
			ExpirationTime: time.Now().Unix() + config.Data.EmailCodeLifetime,
		}
		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if config.Data.EmailEnabled {
			type templateData struct {
				Name string
				Url  string
			}
			body, err := services.ParseEmailTemplate("forgotPassword", c.Get("lang").(string), templateData{
				Name: user.Name,
				Url:  fmt.Sprintf("https://%s/auth/forgotPassword?email=%s&token=%s", config.Data.DomainName, body.Email, code),
			})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			go services.SendEmail([]string{user.Email}, services.Tr("H-Bank Reset Password", lang), body)
		}
		return c.JSON(http.StatusOK, responses.New(true, "An email with a reset password link has been sent to the specified address", lang))
	}

	return c.JSON(http.StatusOK, responses.New(false, "Invalid captcha token", lang))
}

// /v1/auth/resetPassword (POST)
func (h *Handler) ResetPassword(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.ResetPassword
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if !services.IsValidEmail(body.Email) {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid email", lang))
	}

	if len(body.NewPassword) > config.Data.MaxPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf(services.Tr("New password too long (max %d)", lang), config.Data.MaxPasswordLength),
		})
	}

	if utf8.RuneCountInString(body.NewPassword) < config.Data.MinPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf(services.Tr("New password too short (min %d)", lang), config.Data.MinPasswordLength),
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	token, err := h.userStore.GetResetPasswordCode(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if token == nil || subtle.ConstantTimeCompare(token.CodeHash, services.HashToken(body.Token)) == 0 {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}
	h.userStore.DeleteResetPasswordCode(token)
	if token.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(body.NewPassword), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully changed password", lang))
}

// /v1/auth/requestChangeEmail (POST)
func (h *Handler) RequestChangeEmail(c echo.Context) error {
	lang := c.Get("lang").(string)
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.ChangeEmailRequest
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if !services.IsValidEmail(body.NewEmail) {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid new email", lang))
	}

	if services.VerifyCaptcha(body.CaptchaToken) {
		if u, _ := h.userStore.GetByEmail(body.NewEmail); u != nil {
			return c.JSON(http.StatusOK, responses.New(false, "The user with this email does already exist", lang))
		}

		if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
			return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
		}

		emailCode, err := h.userStore.GetChangeEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		h.userStore.DeleteChangeEmailCode(emailCode)
		code := services.GenerateRandomString(64)
		user.ChangeEmailCode = models.ChangeEmailCode{
			CodeHash:       services.HashToken(code),
			ExpirationTime: time.Now().Unix() + config.Data.EmailCodeLifetime,
			NewEmail:       body.NewEmail,
		}
		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if config.Data.EmailEnabled {
			type templateData struct {
				Name string
				Url  string
			}
			emailBody, err := services.ParseEmailTemplate("changeEmail", c.Get("lang").(string), templateData{
				Name: user.Name,
				Url:  fmt.Sprintf("https://%s/auth/changeEmail?token=%s", config.Data.DomainName, code),
			})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			go services.SendEmail([]string{body.NewEmail}, services.Tr("H-Bank Change Email", lang), emailBody)
		}
		return c.JSON(http.StatusOK, responses.New(true, "An email with a change email link has been sent to the new email address", lang))
	}

	return c.JSON(http.StatusOK, responses.New(false, "Invalid captcha token", lang))
}

// /v1/auth/changeEmail (POST)
func (h *Handler) ChangeEmail(c echo.Context) error {
	lang := c.Get("lang").(string)
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.ChangeEmail
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	token, err := h.userStore.GetChangeEmailCode(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if token == nil || subtle.ConstantTimeCompare(token.CodeHash, services.HashToken(body.Token)) == 0 {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}
	h.userStore.DeleteChangeEmailCode(token)
	if token.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	if u, _ := h.userStore.GetByEmail(token.NewEmail); u != nil {
		return c.JSON(http.StatusOK, responses.New(false, "The user with this email does already exist", lang))
	}

	user.Email = token.NewEmail
	user.EmailConfirmed = true

	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewAuthUser(user))
}
