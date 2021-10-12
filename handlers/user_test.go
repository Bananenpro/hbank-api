package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/db"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/router"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetUser(t *testing.T) {
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
		Name:            "bob",
		Email:           "bob@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
	}
	us.Create(user1)

	user2 := &models.User{
		Name:            "peter",
		Email:           "peter@gmail.com",
		EmailConfirmed:  true,
		TwoFaOTPEnabled: true,
	}
	us.Create(user2)

	handler := New(us)

	tests := []struct {
		tName       string
		user        *models.User
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantAllInfo bool
	}{
		{tName: "Doesn't exist", wantCode: http.StatusNotFound, wantSuccess: false, wantMessage: "Resource not found"},
		{tName: "Auth user", user: user1, wantCode: http.StatusOK, wantSuccess: true, wantAllInfo: true},
		{tName: "Not auth user", user: user2, wantCode: http.StatusOK, wantSuccess: true, wantAllInfo: false},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)

			c.SetParamNames("id")
			if tt.user != nil {
				c.SetParamValues(tt.user.Id.String())
			} else {
				c.SetParamValues(uuid.NewString())
			}
			c.Set("userId", user1.Id)

			err := handler.GetUser(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"id":"%s"`, tt.user.Id.String()))
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"name":"%s"`, tt.user.Name))

				if tt.wantAllInfo {
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"email":"%s"`, tt.user.Email))
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"email_confirmed":%t`, tt.user.EmailConfirmed))
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"two_fa_otp_enabled":%t`, tt.user.TwoFaOTPEnabled))
				} else {
					assert.NotContains(t, rec.Body.String(), fmt.Sprintf(`"email"`))
					assert.NotContains(t, rec.Body.String(), fmt.Sprintf(`"email_confirmed"`))
					assert.NotContains(t, rec.Body.String(), fmt.Sprintf(`"two_fa_otp_enabled"`))
				}
			}
		})
	}
}
