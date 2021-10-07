package services

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Bananenpro/hbank2-api/config"
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
