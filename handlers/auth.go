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
	minNameLenght     = 3
	minPasswordLenght = 6
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

	if utf8.RuneCountInString(body.Name) < minNameLenght {
		return c.JSON(http.StatusBadRequest, responses.RegisterInvalid{
			Generic: responses.Generic{
				Message: "Name too short",
			},
			MinNameLength:     minNameLenght,
			MinPasswordLength: minPasswordLenght,
		})
	}

	if utf8.RuneCountInString(body.Password) < minPasswordLenght {
		return c.JSON(http.StatusBadRequest, responses.RegisterInvalid{
			Generic: responses.Generic{
				Message: "Password too short",
			},
			MinNameLength:     minNameLenght,
			MinPasswordLength: minPasswordLenght,
		})
	}

	userId, err := services.Register(c, body.Email, body.Name, body.Password)
	if err != nil {
		switch err {
		case services.ErrAuthEmailExists:
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

func isValidEmail(email string) bool {
	if utf8.RuneCountInString(email) < 3 || utf8.RuneCountInString(email) > 254 {
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
