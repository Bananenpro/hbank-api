package handlers

import (
	"bytes"
	"fmt"
	"image/png"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pquerna/otp/totp"
	"gitlab.com/Bananenpro05/hbank2-api/bindings"
	"gitlab.com/Bananenpro05/hbank2-api/config"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"gitlab.com/Bananenpro05/hbank2-api/responses"
	"gitlab.com/Bananenpro05/hbank2-api/services"
	"golang.org/x/crypto/bcrypt"
)

const (
	minNameLength     = 3
	minPasswordLength = 6
	maxNameLength     = 254
	maxPasswordLength = 254
	maxEmailLength    = 254
)

var (
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
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

	if !isValidEmail(body.Email) {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "Invalid email",
		})
	}

	if len(body.Name) > maxNameLength {
		return c.JSON(http.StatusOK, responses.NewRegisterInvalid("Name too long"))
	}

	if utf8.RuneCountInString(body.Name) < minNameLength {
		return c.JSON(http.StatusOK, responses.NewRegisterInvalid("Name too short"))
	}

	if len(body.Password) > maxPasswordLength {
		return c.JSON(http.StatusOK, responses.NewRegisterInvalid("Password too long"))
	}

	if utf8.RuneCountInString(body.Password) < minPasswordLength {
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

// /v1/auth/confirmEmail?email=string (GET)
func (h *Handler) SendConfirmEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if !isValidEmail(email) {
		return c.JSON(http.StatusBadRequest, responses.Base{
			Success: false,
			Message: "Missing or invalid email query parameter",
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

	if body.LoginToken == "" {
		user, err := h.userStore.GetByEmail(body.Email)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
		}

		if totp.Validate(body.OTPCode, user.OtpSecret) {
			user.TwoFaOTPEnabled = true
			err = h.userStore.Update(user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
			}
			return c.JSON(http.StatusOK, responses.Base{
				Success: true,
				Message: "Correct code",
			})
		}

		return c.JSON(http.StatusUnauthorized, responses.Base{
			Success: false,
			Message: "Invalid credentials",
		})
	} else {
		// TODO: login
		return c.JSON(http.StatusNotImplemented, responses.Base{
			Success: false,
			Message: "Not yet implemented",
		})
	}
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

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials())
	}

	if !user.EmailConfirmed {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "Email is not confirmed",
		})
	}

	if !user.TwoFaOTPEnabled {
		return c.JSON(http.StatusOK, responses.Base{
			Success: false,
			Message: "2FA is not enabled",
		})
	}

	code := ""
	exists := true

	for exists {
		code = services.GenerateRandomString(64)
		t, err := h.userStore.GetLoginTokenByCode(user, code)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
		}
		exists = t != nil
	}

	user.LoginTokens = append(user.LoginTokens, models.LoginToken{
		Code:           code,
		ExpirationTime: time.Now().UnixMilli() + config.Data.LoginTokenLifetime,
	})
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err))
	}

	return c.JSON(http.StatusOK, responses.Login{
		Base: responses.Base{
			Success: true,
			Message: "Successfully signed in",
		},
		LoginToken: code,
	})
}

// Helper functions

func isValidEmail(email string) bool {
	if len(email) > maxEmailLength || utf8.RuneCountInString(email) < 3 {
		return false
	}

	if !emailRegex.MatchString(email) {
		return false
	}

	mx, err := net.LookupMX(strings.Split(email, "@")[1])
	if err != nil || len(mx) == 0 {
		return false
	}

	return true
}
