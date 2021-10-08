package handlers

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Bananenpro/hbank2-api/bindings"
	"github.com/Bananenpro/hbank2-api/config"
	"github.com/Bananenpro/hbank2-api/models"
	"github.com/Bananenpro/hbank2-api/responses"
	"github.com/Bananenpro/hbank2-api/services"
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

	_, err = h.userStore.NewRecoveryCodes(user)
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
	if lastSend+config.Data.SendEmailTimeout > time.Now().UnixMilli() {
		return c.JSON(http.StatusTooManyRequests, responses.Base{
			Success: false,
			Message: fmt.Sprintf("Please wait at least %d minutes between confirm email requests", config.Data.SendEmailTimeout/time.Minute.Milliseconds()),
		})
	}
	h.userStore.SetConfirmEmailLastSent(email, time.Now().UnixMilli())

	user, err := h.userStore.GetByEmail(email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}
	if user != nil {
		emailCode, err := h.userStore.GetEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}

		if !user.EmailConfirmed {
			h.userStore.DeleteEmailCode(emailCode)
			user.EmailCode = models.EmailCode{
				Code:           services.GenerateRandomString(6),
				ExpirationTime: time.Now().UnixMilli() + config.Data.EmailCodeLifetime,
			}
			err = h.userStore.Update(user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
			}

			if config.Data.EmailEnabled {
				type templateData struct {
					Name    string
					Content string
				}
				body, err := services.ParseEmailTemplate("email.html", templateData{
					Name:    user.Name,
					Content: "der Code lautet: " + user.EmailCode.Code,
				})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
				}
				go services.SendEmail([]string{user.Email}, "H-Bank Bestätigungscode", body)
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

	success := false

	if user != nil {
		emailCode, err := h.userStore.GetEmailCode(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		if emailCode != nil {
			if emailCode.Code == body.Code {
				if emailCode.ExpirationTime > time.Now().UnixMilli() {
					user.EmailConfirmed = true
					err = h.userStore.Update(user)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
					}

					success = true
				}

				h.userStore.DeleteEmailCode(emailCode)
			}
		}
	}

	if success {
		return c.JSON(http.StatusOK, responses.Base{
			Success: true,
			Message: "Successfully confirmed email address",
		})
	} else {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "Email was not confirmed",
		})
	}
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

// /v1/auth/twoFactor/otp/verify (POST)
func (h *Handler) VerifyOTPCode(c echo.Context) error {
	var body bindings.VerifyOTPCode
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

	if totp.Validate(body.OTPCode, user.OtpSecret) {
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
			Code:           code,
			ExpirationTime: time.Now().UnixMilli() + config.Data.LoginTokenLifetime,
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

// /v1/auth/login (POST)
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
		Code:           code,
		ExpirationTime: time.Now().UnixMilli() + config.Data.LoginTokenLifetime,
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

	if passwordToken.ExpirationTime < time.Now().UnixMilli() {
		h.userStore.DeletePasswordToken(passwordToken)
	}
	if twoFAToken.ExpirationTime < time.Now().UnixMilli() {
		h.userStore.DeleteTwoFAToken(twoFAToken)
	}
	if passwordToken.ExpirationTime < time.Now().UnixMilli() || twoFAToken.ExpirationTime < time.Now().UnixMilli() {
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

	refreshToken := &models.RefreshToken{
		Code:           services.GenerateRandomString(64),
		ExpirationTime: time.Now().UnixMilli() + config.Data.RefreshTokenLifetime,
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
		Value:    refreshToken.Code,
		MaxAge:   int(config.Data.RefreshTokenLifetime / 1000),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth/refresh",
	})

	c.SetCookie(&http.Cookie{
		Name:     "Auth-Token",
		Value:    authToken,
		MaxAge:   int(config.Data.AuthTokenLifetime / 1000),
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})

	c.SetCookie(&http.Cookie{
		Name:     "Auth-Token-Signature",
		Value:    authTokenSignature,
		MaxAge:   int(config.Data.AuthTokenLifetime / 1000),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return c.JSON(http.StatusOK, responses.Base{
		Success: true,
		Message: "Successfully signed in",
	})
}
