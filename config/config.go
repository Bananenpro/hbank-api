package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

type ConfigData struct {
	Debug                 bool   `json:"debug"`
	ServerPort            int    `json:"server_port"`
	DomainName            string `json:"domain_name"`
	JWTSecret             string `json:"jwt_secret"`
	CaptchaEnabled        bool   `json:"captcha_enabled"`
	CaptchaVerifyUrl      string `json:"captcha_verify_url"`
	CaptchaSecret         string `json:"captcha_secret"`
	CaptchaSiteKey        string `json:"captcha_site_key"`
	EmailEnabled          bool   `json:"email_enabled"`
	EmailHost             string `json:"email_host"`
	EmailPort             int    `json:"email_port"`
	EmailUsername         string `json:"email_username"`
	EmailPassword         string `json:"email_password"`
	UserMinNameLength     int    `json:"user_min_name_length"`
	UserMinPasswordLength int    `json:"user_min_password_length"`
	UserMinEmailLength    int    `json:"user_min_email_length"`
	UserMaxNameLength     int    `json:"user_max_name_length"`
	UserMaxPasswordLength int    `json:"user_max_password_length"`
	UserMaxEmailLength    int    `json:"user_max_email_length"`
	BcryptCost            int    `json:"bcrypt_cost"`
	LoginTokenLifetime    int64  `json:"login_token_lifetime"`
	EmailCodeLifetime     int64  `json:"email_code_lifetime"`
	AuthTokenLifetime     int64  `json:"auth_token_lifetime"`
	RefreshTokenLifetime  int64  `json:"refresh_token_lifetime"`
	SendEmailTimeout      int64  `json:"send_email_timeout"`
}

var defaultData = ConfigData{
	ServerPort:            8080,
	DomainName:            "hbank",
	UserMinNameLength:     3,
	UserMinPasswordLength: 6,
	UserMinEmailLength:    3,
	UserMaxNameLength:     255,
	UserMaxPasswordLength: 255,
	UserMaxEmailLength:    255,
	BcryptCost:            10,
	LoginTokenLifetime:    time.Minute.Milliseconds() * 5,
	EmailCodeLifetime:     time.Minute.Milliseconds() * 5,
	AuthTokenLifetime:     time.Minute.Milliseconds() * 10,
	RefreshTokenLifetime:  31557600000, // 1 year
	SendEmailTimeout:      time.Minute.Milliseconds() * 2,
}

var Data = defaultData

// @param filepaths A slice of config filepaths (json files)
// Will load only from the first valid config file in the list.
func Load(filepaths []string) {
	for _, path := range filepaths {
		if _, err := os.Stat(path); err == nil {
			content, err := os.ReadFile(path)
			if err == nil {
				err = json.Unmarshal(content, &Data)
				if err == nil {
					verifyData()
					return
				} else {
					log.Printf("Couldn't decode config file '%s': %s\n", path, err)
				}
			} else {
				log.Printf("Couldn't read config file '%s': %s\n", path, err)
			}
		}
	}

	log.Println("No config file found")
}

func verifyData() {
	if Data.ServerPort <= 1023 || Data.ServerPort > 65353 {
		log.Println("WARNING: Invalid port number. Using default port: ", defaultData.ServerPort)
		Data.ServerPort = defaultData.ServerPort
	}

	if Data.CaptchaEnabled {
		if Data.CaptchaSecret == "" {
			log.Println("WARNING: No captcha secret specified. Disabling captcha.")
			Data.CaptchaEnabled = false
		}
		if Data.CaptchaSiteKey == "" {
			log.Println("WARNING: No captcha site key specified. Disabling captcha.")
			Data.CaptchaEnabled = false
		}
	}

	if Data.EmailEnabled {
		if Data.EmailHost == "" {
			log.Println("WARNING: No email host provided")
		}

		if Data.EmailPort < 0 || Data.EmailPort > 65353 {
			log.Println("WARNING: Invalid or missing email port")
		}

		if Data.EmailUsername == "" {
			log.Println("WARNING: No email username provided")
		}

		if Data.EmailPassword == "" {
			log.Println("WARNING: No email password provided")
		}
	} else {
		log.Println("WARNING: Email disabled")
	}

	if strings.TrimSpace(Data.DomainName) == "" {
		log.Println("WARNING: Empty domain name. Using default: hbank")
	}

	if len(Data.JWTSecret) < 10 {
		log.Fatalln("ERROR: Please specify a jwt secret (>=10 characters)")
	}
}
