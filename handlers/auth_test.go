package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	e := echo.New()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	models.AutoMigrate(db)
	if err != nil {
		t.Fatal("Unable to connect to in-memory database")
	}

	user := models.User{
		Name:  "bob",
		Email: "exists@gmail.com",
	}
	db.Create(&user)

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
		{tName: "Name too short", name: "bo", email: "bob@gmail.com", password: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Name too short"},
		{tName: "Name too long", name: "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata ",
			email: "bob@gmail.com", password: "123456", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Name too long"},
		{tName: "Password too short", name: "bob", email: "bob@gmail.com", password: "12345", wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Password too short"},
		{tName: "Password too long", name: "bob", email: "bob@gmail.com", password: "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata ",
			wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Password too long"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			// JSON
			jsonBody := fmt.Sprintf(`{"name":"%s","email": "%s","password":"%s"}`, tt.name, tt.email, tt.password)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set(models.DBContextKey, db)

			err := Register(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			// Reset db
			if rec.Code == http.StatusCreated {
				db.Delete(models.User{}, "email = ?", tt.email)
			}

			// FORM
			formValues := make(url.Values)
			formValues.Set("name", tt.name)
			formValues.Set("email", tt.email)
			formValues.Set("password", tt.password)

			req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formValues.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec)
			c.Set(models.DBContextKey, db)

			err = Register(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			// Reset db
			if rec.Code == http.StatusCreated {
				db.Delete(models.User{}, "email = ?", tt.email)
			}
		})
	}
}

func Test_isValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"Valid gmail email", "test@gmail.com", true},
		{"Valid gmx email", "test@gmx.de", true},
		{"Valid outlook email", "test@outlook.com", true},
		{"Valid protonmail email", "test@protonmail.com", true},
		{"Empty string", "", false},
		{"Missing @ sign", "test.gmail.com", false},
		{"Missing name", "@gmail.com", false},
		{"Missing domain", "test@com", false},
		{"Non-existant domain", "test@foomail.abc", false},
		{"Two @ signs", "test@foomail@abc", false},
		{"Too long", "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata ",
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidEmail(tt.email))
		})
	}
}
