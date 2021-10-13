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
	var body bindings.Register
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if !services.VerifyCaptcha(body.CaptchaToken) {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "Invalid captcha token",
		})
	}

	body.Email = strings.ToLower(body.Email)

	if !services.IsValidEmail(body.Email) {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "Invalid email",
		})
	}

	if len(body.Name) > config.Data.UserMaxNameLength {
		return c.JSON(http.StatusOK, responses.NewRegisterInvalid("Name too long"))
	}

	if utf8.RuneCountInString(body.Name) < config.Data.UserMinNameLength {
		return c.JSON(http.StatusOK, responses.NewRegisterInvalid("Name too short"))
	}

	if len(body.Password) > config.Data.UserMaxPasswordLength {
		return c.JSON(http.StatusOK, responses.NewRegisterInvalid("Password too long"))
	}

	if utf8.RuneCountInString(body.Password) < config.Data.UserMinPasswordLength {
		return c.JSON(http.StatusOK, responses.NewRegisterInvalid("Password too short"))
	}

	if u, _ := h.userStore.GetByEmail(body.Email); u != nil {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "The user with this email does already exist",
		})
	}

	user := &models.User{
		Name:             body.Name,
		Email:            body.Email,
		ProfilePictureId: uuid.New(),
	}

	var err error
	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(body.Password), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	err = h.userStore.Create(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	codes, err := h.userStore.NewRecoveryCodes(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusCreated, responses.RegisterSuccess{
		Base: responses.Base{
			Success: true,
			Message: "Successfully registered new user",
		},
		UserId:    user.Id.String(),
		UserEmail: body.Email,
		Codes:     codes,
	})
}

// /v1/auth/confirmEmail/:email (GET)
func (h *Handler) SendConfirmEmail(c echo.Context) error {
	email := c.Param("email")
	if !services.IsValidEmail(email) {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Missing or invalid email parameter",
		})
	}

	lastSend, err := h.userStore.GetConfirmEmailLastSent(email)
	if lastSend+config.Data.SendEmailTimeout > time.Now().Unix() {
		return c.JSON(http.StatusTooManyRequests, responses.Base{
			Success: false,
			Message: fmt.Sprintf("Please wait at least %d minutes between confirm email requests", config.Data.SendEmailTimeout/60),
		})
	}
	h.userStore.SetConfirmEmailLastSent(email, time.Now().Unix())

	user, err := h.userStore.GetByEmail(email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user != nil {
		emailCode, err := h.userStore.GetConfirmEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		if !user.EmailConfirmed {
			h.userStore.DeleteConfirmEmailCode(emailCode)
			code := services.GenerateRandomString(6)
			user.ConfirmEmailCode = models.ConfirmEmailCode{
				CodeHash: services.HashToken(code),
			}
			err = h.userStore.Update(user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
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
					return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
				}
				go services.SendEmail([]string{user.Email}, "H-Bank Confirm Email", body)
			}
		}
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "If the email address is linked to a user whose email has not yet been confirmed, a code has been sent to the specified address",
	})
}

// /v1/auth/confirmEmail (POST)
func (h *Handler) VerifyConfirmEmailCode(c echo.Context) error {
	var body bindings.ConfirmEmail
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	if user != nil {
		confirmEmailCode, err := h.userStore.GetConfirmEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		if confirmEmailCode != nil {
			if subtle.ConstantTimeCompare(confirmEmailCode.CodeHash, services.HashToken(body.Code)) == 1 {
				user.EmailConfirmed = true
				err = h.userStore.Update(user)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
				}

				h.userStore.DeleteConfirmEmailCode(confirmEmailCode)
				return c.JSON(http.StatusOK, responses.Base{
					Success: true,
					Message: "Successfully confirmed email address",
				})
			}
		}
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: false,
		Message: "Email was not confirmed",
	})
}

// /v1/auth/twoFactor/otp/activate (POST)
func (h *Handler) Activate2FAOTP(c echo.Context) error {
	var body bindings.Activate2FAOTP
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	if !user.TwoFaOTPEnabled {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      config.Data.DomainName,
			AccountName: user.Email,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		img, err := key.Image(200, 200)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		var qr bytes.Buffer

		png.Encode(&qr, img)

		secret := key.Secret()

		user.OtpSecret = secret
		user.OtpQrCode = qr.Bytes()

		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		return c.Blob(http.StatusOK, "image/png", user.OtpQrCode)
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: false,
		Message: "TwoFaOTP is already activated",
	})
}

// /v1/auth/twoFactor/otp/get (POST)
func (h *Handler) GetOTPQRCode(c echo.Context) error {
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "The user does no longer exist",
		})
	}

	var body bindings.Password
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}

	return c.Blob(http.StatusOK, "image/png", user.OtpQrCode)
}

// /v1/auth/twoFactor/otp/verify (POST)
func (h *Handler) VerifyOTPCode(c echo.Context) error {
	var body bindings.VerifyCode
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	if totp.Validate(body.Code, user.OtpSecret) {
		user.TwoFaOTPEnabled = true
		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		code := ""
		exists := true

		for exists {
			code = services.GenerateRandomString(64)
			t, err := h.userStore.GetTwoFATokenByCode(user, code)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
			}
			exists = t != nil
		}

		user.TwoFATokens = append(user.TwoFATokens, models.TwoFAToken{
			CodeHash:       services.HashToken(code),
			ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime,
		})
		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		return c.JSON(http.StatusOK, responses.Token{
			Base: responses.Base{
				Success: true,
				Message: "Successfully aquired two factor token",
			},
			Token: code,
		})
	}

	return c.JSON(http.StatusUnauthorized, responses.Base{
		Success: false,
		Message: "Invalid credentials",
	})
}

// /v1/auth/twoFactor/otp/new
func (h *Handler) NewOTP(c echo.Context) error {
	var body bindings.Password
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "The user does no longer exist",
		})
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}

	if user.TwoFaOTPEnabled {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      config.Data.DomainName,
			AccountName: user.Email,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		img, err := key.Image(200, 200)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		var qr bytes.Buffer

		png.Encode(&qr, img)

		secret := key.Secret()

		user.OtpSecret = secret
		user.OtpQrCode = qr.Bytes()

		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		return c.Blob(http.StatusOK, "image/png", user.OtpQrCode)
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: false,
		Message: "Please enable otp first",
	})
}

// /v1/auth/passwordAuth (POST)
func (h *Handler) PasswordAuth(c echo.Context) error {
	var body bindings.PasswordAuth
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	code := ""
	exists := true

	for exists {
		code = services.GenerateRandomString(64)
		t, err := h.userStore.GetPasswordTokenByCode(user, code)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		exists = t != nil
	}

	user.PasswordTokens = append(user.PasswordTokens, models.PasswordToken{
		CodeHash:       services.HashToken(code),
		ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime,
	})
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusOK, responses.Token{
		Base: responses.Base{
			Success: true,
			Message: "Successfully aquired password token",
		},
		Token: code,
	})
}

// /v1/auth/login (POST)
func (h *Handler) Login(c echo.Context) error {
	var body bindings.Login
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	passwordToken, err := h.userStore.GetPasswordTokenByCode(user, body.PasswordToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if passwordToken == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if twoFAToken == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	if passwordToken.ExpirationTime < time.Now().Unix() {
		h.userStore.DeletePasswordToken(passwordToken)
	}
	if twoFAToken.ExpirationTime < time.Now().Unix() {
		h.userStore.DeleteTwoFAToken(twoFAToken)
	}
	if passwordToken.ExpirationTime < time.Now().Unix() || twoFAToken.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	h.userStore.DeletePasswordToken(passwordToken)
	h.userStore.DeleteTwoFAToken(twoFAToken)

	if !user.EmailConfirmed {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "Email is not confirmed",
		})
	}

	if !user.TwoFaOTPEnabled {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "Two factor authentication is not enabled",
		})
	}

	code := services.GenerateRandomString(64)
	hash, err := bcrypt.GenerateFromPassword([]byte(code), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	refreshToken := &models.RefreshToken{
		CodeHash:       hash,
		ExpirationTime: time.Now().Unix() + config.Data.RefreshTokenLifetime,
	}
	err = h.userStore.AddRefreshToken(user, refreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	authToken, authTokenSignature, err := services.NewAuthToken(user)
	if err != nil {
		h.userStore.DeleteRefreshToken(refreshToken)
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
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

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "Successfully signed in",
	})
}

// /v1/auth/refresh (POST)
func (h *Handler) Refresh(c echo.Context) error {
	var body bindings.Refresh
	err := c.Bind(&body)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	userId, err := uuid.Parse(body.UserId)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	refreshCookie, err := c.Cookie("Refresh-Token")
	if err != nil || len(refreshCookie.Value) <= 36 {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	refreshTokenId, err := uuid.Parse(refreshCookie.Value[:36])
	if err != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	refreshToken, err := h.userStore.GetRefreshToken(user, refreshTokenId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if refreshToken == nil || bcrypt.CompareHashAndPassword(refreshToken.CodeHash, []byte(refreshCookie.Value[36:])) != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	if refreshToken.Used {
		err = h.userStore.DeleteRefreshTokens(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	newRefreshToken, code, err := h.userStore.RotateRefreshToken(user, refreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	authToken, authTokenSignature, err := services.NewAuthToken(user)
	if err != nil {
		h.userStore.DeleteRefreshToken(newRefreshToken)
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
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

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "Successfully refreshed tokens",
	})
}

// /v1/auth/logout?all=bool (POST)
func (h *Handler) Logout(c echo.Context) error {
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "The user does no longer exist",
		})
	}

	if services.StrToBool(c.QueryParams().Get("all")) {
		err = h.userStore.DeleteRefreshTokens(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
	} else {
		refreshCookie, err := c.Cookie("Refresh-Token")
		if err != nil || len(refreshCookie.Value) <= 36 {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		tokenId, err := uuid.Parse(refreshCookie.Value[:36])
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		refreshToken, err := h.userStore.GetRefreshToken(user, tokenId)
		if err != nil || refreshToken == nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		err = h.userStore.DeleteRefreshToken(refreshToken)
		if err != nil || refreshToken == nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "Successfully signed out",
	})
}

// /v1/auth/twoFactor/recovery/verify (POST)
func (h *Handler) VerifyRecoveryCode(c echo.Context) error {
	var body bindings.VerifyCode
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	code, err := h.userStore.GetRecoveryCodeByCode(user, body.Code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if code == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	token := ""
	exists := true

	for exists {
		token = services.GenerateRandomString(64)
		t, err := h.userStore.GetTwoFATokenByCode(user, token)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		exists = t != nil
	}

	err = h.userStore.DeleteRecoveryCode(code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	user.TwoFATokens = append(user.TwoFATokens, models.TwoFAToken{
		CodeHash:       services.HashToken(token),
		ExpirationTime: time.Now().Unix() + config.Data.LoginTokenLifetime,
	})
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusOK, responses.Token{
		Base: responses.Base{
			Success: true,
			Message: "Successfully aquired two factor token",
		},
		Token: token,
	})
}

// /v1/auth/twoFactor/recovery/new (POST)
func (h *Handler) NewRecoveryCodes(c echo.Context) error {
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "The user does no longer exist",
		})
	}

	var body bindings.Password
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}

	codes, err := h.userStore.NewRecoveryCodes(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
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
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "The user does no longer exist",
		})
	}

	var body bindings.ChangePassword
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if len(body.NewPassword) > config.Data.UserMaxPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf("New password too long (max %d)", config.Data.UserMaxPasswordLength),
		})
	}

	if utf8.RuneCountInString(body.NewPassword) < config.Data.UserMinPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf("New password too short (min %d)", config.Data.UserMinPasswordLength),
		})
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.OldPassword)) != nil {
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

	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(body.NewPassword), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "Successfully changed password",
	})
}

// /v1/auth/forgotPassword (POST)
func (h *Handler) ForgotPassword(c echo.Context) error {
	var body bindings.ForgotPassword
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if !services.IsValidEmail(body.Email) {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid email",
		})
	}

	if services.VerifyCaptcha(body.CaptchaToken) {
		user, err := h.userStore.GetByEmail(body.Email)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		if user == nil {
			return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
		}
		twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		if twoFAToken == nil {
			return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
		}
		h.userStore.DeleteTwoFAToken(twoFAToken)
		if twoFAToken.ExpirationTime < time.Now().Unix() {
			return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
		}

		lastSend, err := h.userStore.GetForgotPasswordEmailLastSent(body.Email)
		if lastSend+config.Data.SendEmailTimeout > time.Now().Unix() {
			return c.JSON(http.StatusTooManyRequests, responses.Base{
				Success: false,
				Message: fmt.Sprintf("Please wait at least %d minutes between forgot password email requests", config.Data.SendEmailTimeout/60),
			})
		}
		h.userStore.SetForgotPasswordEmailLastSent(body.Email, time.Now().Unix())

		emailCode, err := h.userStore.GetResetPasswordCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		h.userStore.DeleteResetPasswordCode(emailCode)
		code := services.GenerateRandomString(64)
		user.ResetPasswordCode = models.ResetPasswordCode{
			CodeHash:       services.HashToken(code),
			ExpirationTime: time.Now().Unix() + config.Data.EmailCodeLifetime,
		}
		err = h.userStore.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
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
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
			}
			go services.SendEmail([]string{user.Email}, "H-Bank Reset Password", body)
		}
		return c.JSON(http.StatusOK, responses.Base{
			Success: true,
			Message: "An email with a reset password link has been sent to the specified address",
		})
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: false,
		Message: "Invalid captcha token",
	})
}

// /v1/auth/resetPassword (POST)
func (h *Handler) ResetPassword(c echo.Context) error {
	var body bindings.ResetPassword
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if !services.IsValidEmail(body.Email) {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid email",
		})
	}

	if len(body.NewPassword) > config.Data.UserMaxPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf("New password too long (max %d)", config.Data.UserMaxPasswordLength),
		})
	}

	if utf8.RuneCountInString(body.NewPassword) < config.Data.UserMinPasswordLength {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: fmt.Sprintf("New password too short (min %d)", config.Data.UserMinPasswordLength),
		})
	}

	user, err := h.userStore.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	token, err := h.userStore.GetResetPasswordCode(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if token == nil || subtle.ConstantTimeCompare(token.CodeHash, services.HashToken(body.Token)) == 0 {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}
	h.userStore.DeleteResetPasswordCode(token)
	if token.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(body.NewPassword), config.Data.BcryptCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "Successfully changed password",
	})
}

// /v1/auth/requestChangeEmail (POST)
func (h *Handler) RequestChangeEmail(c echo.Context) error {
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "The user does no longer exist",
		})
	}

	var body bindings.ChangeEmailRequest
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if !services.IsValidEmail(body.NewEmail) {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid new email",
		})
	}

	if services.VerifyCaptcha(body.CaptchaToken) {
		if u, _ := h.userStore.GetByEmail(body.NewEmail); u != nil {
			return c.JSON(http.StatusOK, responses.Base{
				Success: false,
				Message: "The user with this email does already exist",
			})
		}

		if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
			fmt.Println("password")
			return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
		}

		twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		if twoFAToken == nil {
			fmt.Println("token")
			return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
		}
		h.userStore.DeleteTwoFAToken(twoFAToken)
		if twoFAToken.ExpirationTime < time.Now().Unix() {
			fmt.Println("expired")
			return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
		}

		emailCode, err := h.userStore.GetChangeEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
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
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
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
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
			}
			go services.SendEmail([]string{body.NewEmail}, "H-Bank Change Email", emailBody)
		}
		return c.JSON(http.StatusOK, responses.Base{
			Success: true,
			Message: "An email with a change email link has been sent to the new email address",
		})
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: false,
		Message: "Invalid captcha token",
	})
}

// /v1/auth/changeEmail (POST)
func (h *Handler) ChangeEmail(c echo.Context) error {
	user, err := h.userStore.GetById(c.Get("userId").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "The user does no longer exist",
		})
	}

	var body bindings.ChangeEmail
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Invalid request body",
		})
	}

	token, err := h.userStore.GetChangeEmailCode(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if token == nil || subtle.ConstantTimeCompare(token.CodeHash, services.HashToken(body.Token)) == 0 {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}
	h.userStore.DeleteChangeEmailCode(token)
	if token.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials())
	}

	if u, _ := h.userStore.GetByEmail(token.NewEmail); u != nil {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "The user with this email does already exist",
		})
	}

	user.Email = token.NewEmail
	user.EmailConfirmed = true

	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "Successfully changed email address",
	})
}
