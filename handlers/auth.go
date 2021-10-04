package handlers

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/bindings"
	"gitlab.com/Bananenpro05/hbank2-api/responses"
	"gitlab.com/Bananenpro05/hbank2-api/services"
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
func Register(c echo.Context) error {
	var body bindings.Register
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, responses.Generic{
			Message: "Invalid request body",
		})
	}

	if err := services.VerifyCaptcha(body.CaptchaToken); err != nil {
		switch err {
		case services.ErrInvalidCaptchaToken:
			return c.JSON(http.StatusForbidden, responses.Generic{
				Message: "Invalid captcha token",
			})
		default:
			return c.JSON(http.StatusInternalServerError, responses.Generic{
				Message: "Due to an unexpected error the user couldn't be registered",
			})
		}
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	if !isValidEmail(body.Email) {
		return c.JSON(http.StatusBadRequest, responses.Generic{
			Message: "Invalid email",
		})
	}

	if len(body.Name) > maxNameLength {
		return c.JSON(http.StatusBadRequest, responses.RegisterInvalid{
			Generic: responses.Generic{
				Message: "Name too long",
			},
			MinNameLength:     minNameLength,
			MinPasswordLength: minPasswordLength,
			MaxNameLength:     maxNameLength,
			MaxPasswordLength: maxPasswordLength,
		})
	}

	if utf8.RuneCountInString(body.Name) < minNameLength {
		return c.JSON(http.StatusBadRequest, responses.RegisterInvalid{
			Generic: responses.Generic{
				Message: "Name too short",
			},
			MinNameLength:     minNameLength,
			MinPasswordLength: minPasswordLength,
			MaxNameLength:     maxNameLength,
			MaxPasswordLength: maxPasswordLength,
		})
	}

	if len(body.Password) > maxPasswordLength {
		return c.JSON(http.StatusBadRequest, responses.RegisterInvalid{
			Generic: responses.Generic{
				Message: "Password too long",
			},
			MinNameLength:     minNameLength,
			MinPasswordLength: minPasswordLength,
			MaxNameLength:     maxNameLength,
			MaxPasswordLength: maxPasswordLength,
		})
	}

	if utf8.RuneCountInString(body.Password) < minPasswordLength {
		return c.JSON(http.StatusBadRequest, responses.RegisterInvalid{
			Generic: responses.Generic{
				Message: "Password too short",
			},
			MinNameLength:     minNameLength,
			MinPasswordLength: minPasswordLength,
			MaxNameLength:     maxNameLength,
			MaxPasswordLength: maxPasswordLength,
		})
	}

	userId, err := services.Register(c, body.Email, body.Name, body.Password)
	if err != nil {
		switch err {
		case services.ErrEmailExists:
			return c.JSON(http.StatusForbidden, responses.Generic{
				Message: "The user with this email does already exist",
			})
		default:
			return c.JSON(http.StatusInternalServerError, responses.Generic{
				Message: "Due to an unexpected error the user couldn't be registered",
			})
		}
	}

	return c.JSON(http.StatusCreated, responses.RegisterSuccess{
		Generic: responses.Generic{
			Message: "Successfully registered new user",
		},
		UserId:    userId.String(),
		UserEmail: body.Email,
	})
}

// /v1/auth/confirmEmail?email=string (GET)
func SendConfirmEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if !isValidEmail(email) {
		return c.JSON(http.StatusBadRequest, responses.Generic{
			Message: "Missing or invalid email query parameter",
		})
	}

	err := services.SendConfirmEmail(c, email)
	if err == services.ErrTimeout {
		return c.JSON(http.StatusTooManyRequests, responses.Generic{
			Message: "Please wait at least 2 minutes between confirm email requests",
		})
	} else if err != nil && err != services.ErrNotFound && err != services.ErrEmailAlreadyConfirmed {
		return c.JSON(http.StatusInternalServerError, responses.Generic{
			Message: "An unexpected error occurred: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, responses.Generic{
		Message: "If the email address is linked to a user whose email has not yet been confirmed, a code has been sent to the specified address",
	})
}

// /v1/auth/confirmEmail (POST)
func VerifyConfirmEmailCode(c echo.Context) error {
	var body bindings.ConfirmEmail
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Generic{
			Message: "Invalid request body",
		})
	}

	if services.VerifyConfirmEmailCode(c, body.Email, body.Code) {
		return c.JSON(http.StatusOK, responses.Generic{
			Message: "Successfully confirmed email address",
		})
	} else {
		return c.JSON(http.StatusForbidden, responses.Generic{
			Message: "Email was not confirmed",
		})
	}
}

// /v1/auth/twoFactor/otp/activate (POST)
func Activate2FAOTP(c echo.Context) error {
	var body bindings.Activate2FAOTP
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Generic{
			Message: "Invalid request body",
		})
	}

	qr, err := services.Activate2FAOTP(c, body.Email, body.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			return c.JSON(http.StatusUnauthorized, responses.Generic{
				Message: "Invalid credentials",
			})
		default:
			return c.JSON(http.StatusInternalServerError, responses.Generic{
				Message: "An unexpected error occurred",
			})
		}
	}

	return c.Blob(http.StatusOK, "image/png", qr)
}

// /v1/auth/twoFactor/otp/verify (POST)
func VerifyOTPCode(c echo.Context) error {
	var body bindings.VerifyOTPCode
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.Generic{
			Message: "Invalid request body",
		})
	}

	if body.LoginToken == "" {
		if services.VerifyOTPCode(c, body.Email, body.OTPCode) {
			return c.JSON(http.StatusOK, responses.Generic{
				Message: "Correct code",
			})
		} else {
			return c.JSON(http.StatusUnauthorized, responses.Generic{
				Message: "Invalid credentials",
			})
		}
	} else {
		// TODO: login
		return c.JSON(http.StatusInternalServerError, responses.Generic{
			Message: "Not yet implemented",
		})
	}
}

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
