package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type ConfigData struct {
	Debug                     bool   `json:"debug"`
	DBVerbose                 bool   `json:"db_verbose"`
	ServerPort                int    `json:"server_port"`
	SSLCertPath               string `json:"ssl_cert_path"`
	SSLKeyPath                string `json:"ssl_key_path"`
	DomainName                string `json:"domain_name"`
	JWTSecret                 string `json:"jwt_secret"`
	CaptchaEnabled            bool   `json:"captcha_enabled"`
	CaptchaVerifyUrl          string `json:"captcha_verify_url"`
	CaptchaSecret             string `json:"captcha_secret"`
	CaptchaSiteKey            string `json:"captcha_site_key"`
	EmailEnabled              bool   `json:"email_enabled"`
	EmailHost                 string `json:"email_host"`
	EmailPort                 int    `json:"email_port"`
	EmailUsername             string `json:"email_username"`
	EmailPassword             string `json:"email_password"`
	MinNameLength             int    `json:"min_name_length"`
	MaxNameLength             int    `json:"max_name_length"`
	MinDescriptionLength      int    `json:"min_description_length"`
	MaxDescriptionLength      int    `json:"max_description_length"`
	MinPasswordLength         int    `json:"min_password_length"`
	MaxPasswordLength         int    `json:"max_password_length"`
	MinEmailLength            int    `json:"min_email_length"`
	MaxEmailLength            int    `json:"max_email_length"`
	MaxProfilePictureFileSize int64  `json:"max_profile_picture_file_size"`
	ProfilePictureSize        int    `json:"profile_picture_size"`
	BcryptCost                int    `json:"bcrypt_cost"`
	PBKDF2Iterations          int    `json:"pbkdf2_iterations"`
	LoginTokenLifetime        int64  `json:"login_token_lifetime"`
	EmailCodeLifetime         int64  `json:"email_code_lifetime"`
	AuthTokenLifetime         int64  `json:"auth_token_lifetime"`
	RefreshTokenLifetime      int64  `json:"refresh_token_lifetime"`
	SendEmailTimeout          int64  `json:"send_email_timeout"`
}

var defaultData = ConfigData{
	ServerPort:                8080,
	DomainName:                "hbank",
	MinNameLength:             3,
	MaxNameLength:             15,
	MinDescriptionLength:      0,
	MaxDescriptionLength:      256,
	MinPasswordLength:         6,
	MinEmailLength:            3,
	MaxPasswordLength:         64,
	MaxEmailLength:            64,
	MaxProfilePictureFileSize: 10000000, // 10 MB
	ProfilePictureSize:        500,
	BcryptCost:                10,
	PBKDF2Iterations:          10000,
	LoginTokenLifetime:        5 * 60,
	EmailCodeLifetime:         5 * 60,
	AuthTokenLifetime:         10 * 60,
	RefreshTokenLifetime:      1 * 365 * 24 * 60 * 60,
	SendEmailTimeout:          2 * 60,
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

	if f, err := os.Open(Data.SSLCertPath); os.IsNotExist(err) {
		log.Fatalf("ERROR: Cannot find ssl cert file `%s`\n", Data.SSLCertPath)
	} else {
		f.Close()
	}

	if f, err := os.Open(Data.SSLKeyPath); os.IsNotExist(err) {
		log.Fatalf("ERROR: Cannot find ssl key file `%s`\n", Data.SSLCertPath)
	} else {
		f.Close()
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
