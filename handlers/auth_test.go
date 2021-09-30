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
		wantMessage string
	}{
		{tName: "Successful register", name: "bob", email: "bob@gmail.com", password: "123456", wantCode: http.StatusCreated, wantMessage: "Successfully registered new user"},
		{tName: "User does already exist", name: "bob", email: "exists@gmail.com", password: "123456", wantCode: http.StatusForbidden, wantMessage: "The user with this email does already exist"},
		{tName: "Name too short", name: "bo", email: "bob@gmail.com", password: "123456", wantCode: http.StatusBadRequest, wantMessage: "Name too short"},
		{tName: "Password too short", name: "bob", email: "bob@gmail.com", password: "12345", wantCode: http.StatusBadRequest, wantMessage: "Password too short"},
		{tName: "Invalid email (wrong format)", name: "bob", email: "bob.gmail.com", password: "123456", wantCode: http.StatusBadRequest, wantMessage: "Invalid email"},
		{tName: "Invalid email (no such provider)", name: "bob", email: "bob@bla.bla", password: "123456", wantCode: http.StatusBadRequest, wantMessage: "Invalid email"},
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
			assert.Contains(t, rec.Body.String(), `"message":"`+tt.wantMessage+`"`)

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
			assert.Contains(t, rec.Body.String(), `"message":"`+tt.wantMessage+`"`)
		})
	}
}
