package services

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Bananenpro/hbank2-api/config"
	"github.com/Bananenpro/hbank2-api/models"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
)

var (
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func VerifyCaptcha(token string) bool {
	if config.Data.CaptchaEnabled {
		formValues := make(url.Values)
		formValues.Set("secret", config.Data.CaptchaSecret)
		formValues.Set("response", token)
		formValues.Set("sitekey", config.Data.CaptchaSiteKey)
		resp, err := http.PostForm(config.Data.CaptchaVerifyUrl, formValues)
		if err != nil {
			log.Printf("Failed to contact '%s': %s\n", config.Data.CaptchaVerifyUrl, err)
			return false
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read verify captcha response: ", err)
			return false
		}

		type Response struct {
			Success bool
		}
		var jsonResp Response
		json.Unmarshal(body, &jsonResp)

		return jsonResp.Success
	}
	return true
}

func init() {
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

func GenerateRandomString(length int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}

func IsValidEmail(email string) bool {
	if len(email) > config.Data.UserMaxEmailLength || utf8.RuneCountInString(email) < config.Data.UserMinEmailLength {
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

type jwtClaims struct {
	jwt.StandardClaims
	UserId string `json:"user_id"`
}

func NewAuthToken(user *models.User) (string, string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + config.Data.AuthTokenLifetime,
			IssuedAt:  time.Now().Unix(),
			Subject:   user.Name,
		},
		UserId: user.Id.String(),
	})

	str, err := token.SignedString([]byte(config.Data.JWTSecret))
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(str, ".")
	if len(parts) != 3 {
		return "", "", errors.New("Generated jwt is not a valid jwt")
	}

	_, valid := VerifyAuthToken(str)
	if !valid {
		return "", "", errors.New("Generated jwt is not a valid jwt")
	}

	return parts[0] + "." + parts[1], parts[2], nil
}

func VerifyAuthToken(authToken string) (uuid.UUID, bool) {
	var claims jwtClaims
	token, err := jwt.ParseWithClaims(authToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.Data.JWTSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, false
	}

	if !token.Valid {
		return uuid.UUID{}, false
	}

	userId, err := uuid.Parse(claims.UserId)
	return userId, err == nil
}

func HashToken(token string) []byte {
	return pbkdf2.Key([]byte(token), []byte(""), 10000, 64, sha512.New)
}
