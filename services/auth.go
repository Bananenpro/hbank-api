package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/config"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	bcryptCost              = 10
	hCaptchaVerifyUrl       = "https://hcaptcha.com/siteverify"
	emailCodeLifetimeMillis = 300000 // 5 min
)

func Register(ctx echo.Context, email, name, password string) (uuid.UUID, error) {
	db := dbFromCtx(ctx)

	if err := db.First(&models.User{}, "email = ?", email).Error; err != gorm.ErrRecordNotFound {
		return uuid.UUID{}, ErrEmailExists
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

func VerifyCaptcha(token string) error {
	if config.Data.CaptchaEnabled {
		formValues := make(url.Values)
		formValues.Set("secret", config.Data.HCaptchaSecret)
		formValues.Set("response", token)
		formValues.Set("sitekey", config.Data.HCaptchaSiteKey)
		resp, err := http.PostForm(hCaptchaVerifyUrl, formValues)
		if err != nil {
			log.Printf("Failed to contact '%s': %s\n", hCaptchaVerifyUrl, err)
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read verify captcha response: ", err)
			return err
		}

		type Response struct {
			Success bool
		}
		var jsonResp Response
		json.Unmarshal(body, &jsonResp)

		if jsonResp.Success {
			return nil
		} else {
			return ErrInvalidCaptchaToken
		}
	}
	return nil
}

func SendConfirmEmail(ctx echo.Context, email string) error {
	db := dbFromCtx(ctx)

	var user models.User
	err := db.Joins("EmailCode").First(&user, "email = ?", email).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return ErrNotFound
		default:
			return err
		}
	}
	if !user.EmailConfirmed {
		code, err := generateRandomString(6)
		if err != nil {
			return err
		}

		db.Delete(&user.EmailCode)
		err = db.Model(&user).Association("EmailCode").Replace(&models.EmailCode{
			Code:           code,
			ExpirationTime: time.Now().UnixMilli() + emailCodeLifetimeMillis,
		})
		if err != nil {
			return err
		}

		if config.Data.EmailEnabled {
			type templateData struct {
				Name    string
				Content string
			}
			body, err := ParseEmailTemplate("template.html", templateData{
				Name:    user.Name,
				Content: "der Code lautet: " + user.EmailCode.Code,
			})
			if err != nil {
				return err
			}
			go SendEmail([]string{user.Email}, "H-Bank Bestätigungscode", body)
		}

		return nil
	} else {
		return ErrEmailAlreadyConfirmed
	}
}

func VerifyConfirmEmailCode(ctx echo.Context, email string, code string) bool {
	db := dbFromCtx(ctx)

	var user models.User
	err := db.Joins("EmailCode").First(&user, "email = ?", email).Error
	if err != nil {
		return false
	}

	success := false

	if user.EmailCode.Code == code {
		if user.EmailCode.ExpirationTime > time.Now().UnixMilli() {
			user.EmailConfirmed = true
			db.Model(&user).Select("email_confirmed").Updates(&user)

			success = true
		}

		db.Delete(&user.EmailCode)
	}

	return success
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
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
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
