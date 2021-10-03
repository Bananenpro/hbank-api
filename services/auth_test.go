package services

import (
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ContextMock struct {
	echo.Context
	db *gorm.DB
}

func (c ContextMock) Get(key string) interface{} {
	if key == models.DBContextKey {
		return c.db
	}
	return nil
}

func TestSendConfirmEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	models.AutoMigrate(db)
	if err != nil {
		t.Fatal("Unable to connect to in-memory database")
	}

	db.Create(&models.User{
		Name:           "confirmed",
		Email:          "confirmed@gmail.com",
		EmailConfirmed: true,
	})

	db.Create(&models.User{
		Name:  "not-confirmed",
		Email: "not.confirmed@gmail.com",
	})

	ctx := ContextMock{
		db: db,
	}

	tests := []struct {
		testName string
		email    string
		wantErr  error
	}{
		{testName: "Existing not yet confirmed user", email: "not.confirmed@gmail.com", wantErr: nil},
		{testName: "Existing already confirmed user", email: "confirmed@gmail.com", wantErr: ErrEmailAlreadyConfirmed},
		{testName: "Non-existing user", email: "doesnt-exist@gmail.com", wantErr: ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, SendConfirmEmail(ctx, tt.email))
		})
	}
}

func TestVerifyConfirmEmailCode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	models.AutoMigrate(db)
	if err != nil {
		t.Fatal("Unable to connect to in-memory database")
	}

	ctx := ContextMock{
		db: db,
	}

	tests := []struct {
		testName string
		email    string
		code     string
		user     models.User
		want     bool
	}{
		{testName: "Correct", email: "bob@gmail.com", code: "abcdef", user: models.User{Email: "bob@gmail.com", EmailCode: models.EmailCode{Code: "abcdef", ExpirationTime: time.Now().UnixMilli() + 10000}}, want: true},
		{testName: "Correct code but expired", email: "bob@gmail.com", code: "abcdef", user: models.User{Email: "bob@gmail.com", EmailCode: models.EmailCode{Code: "abcdef", ExpirationTime: time.Now().UnixMilli() - 10000}}, want: false},
		{testName: "Incorrect code", email: "bob@gmail.com", code: "fedcba", user: models.User{Email: "bob@gmail.com", EmailCode: models.EmailCode{Code: "abcdef", ExpirationTime: time.Now().UnixMilli() + 10000}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			db.Create(&tt.user)

			assert.Equal(t, tt.want, VerifyConfirmEmailCode(ctx, tt.email, tt.code))

			db.Delete(&tt.user)
		})
	}
}
