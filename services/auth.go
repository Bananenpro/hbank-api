package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	bcryptCost = 10
)

var (
	ErrAuthEmailExists = errors.New("email-exists")
)

func Register(ctx echo.Context, email, name, password string) (uuid.UUID, error) {
	db := dbFromCtx(ctx)

	if err := db.First(&models.User{}, "email = ?", email).Error; err != gorm.ErrRecordNotFound {
		return uuid.UUID{}, ErrAuthEmailExists
	}

	user := models.User{
		Name:             name,
		Email:            email,
		ProfilePictureId: uuid.New(),
	}

	var err error
	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return uuid.UUID{}, err
	}

	user.RecoveryCodes, err = generateRecoveryCodes()
	if err != nil {
		return uuid.UUID{}, err
	}

	db.Create(&user)

	return user.Id, nil
}

// ========== Helper functions ==========

func generateRecoveryCodes() ([]models.RecoveryCode, error) {
	codes := make([]models.RecoveryCode, 10)

	var err error
	for i := range codes {
		codes[i].Code, err = generateRandomStringURLSafe(64)
		if err != nil {
			return []models.RecoveryCode{}, errors.New("Couldn't generate recovery codes")
		}
	}

	return codes, nil
}

func dbFromCtx(ctx echo.Context) *gorm.DB {
	return ctx.Get(models.DBContextKey).(*gorm.DB)
}

func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

func generateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomString(length int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func generateRandomStringURLSafe(length int) (string, error) {
	b, err := generateRandomBytes(length)
	return base64.URLEncoding.EncodeToString(b), err
}
