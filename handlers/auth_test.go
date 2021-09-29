package handlers

import (
	"net/http"
	"net/http/httptest"
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
		name        string
		jsonBody    string
		wantCode    int
		wantMessage string
	}{
		{name: "Successful register", jsonBody: `{"name": "bob", "email": "bob@gmail.com", "password": "123456"}`, wantCode: http.StatusCreated, wantMessage: "Successfully registered new user"},
		{name: "User does already exist", jsonBody: `{"name": "bob", "email": "exists@gmail.com", "password": "123456"}`, wantCode: http.StatusForbidden, wantMessage: "The user with this email does already exist"},
		{name: "Name too short", jsonBody: `{"name": "bo", "email": "bob@gmail.com", "password": "123456"}`, wantCode: http.StatusBadRequest, wantMessage: "Name too short"},
		{name: "Password too short", jsonBody: `{"name": "bob", "email": "bob@gmail.com", "password": "12345"}`, wantCode: http.StatusBadRequest, wantMessage: "Password too short"},
		{name: "Invalid email (wrong format)", jsonBody: `{"name": "bob", "email": "bob.gmail.com", "password": "123456"}`, wantCode: http.StatusBadRequest, wantMessage: "Invalid email"},
		{name: "Invalid email (no such provider)", jsonBody: `{"name": "bob", "email": "bob@bla.bla", "password": "123456"}`, wantCode: http.StatusBadRequest, wantMessage: "Invalid email"},
		{name: "Invalid request body", jsonBody: `hehe`, wantCode: http.StatusBadRequest, wantMessage: "Invalid request body"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set(models.DBContextKey, db)

			err := Register(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), `"message":"`+tt.wantMessage+`"`)
		})
	}
}
