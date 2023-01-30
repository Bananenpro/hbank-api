package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type ConfigData struct {
	Debug                     bool   `json:"debug"`
	DBVerbose                 bool   `json:"dbVerbose"`
	ServerPort                int    `json:"serverPort"`
	SSL                       bool   `json:"ssl"`
	SSLCertPath               string `json:"sslCertPath"`
	SSLKeyPath                string `json:"sslKeyPath"`
	DomainName                string `json:"domainName"`
	BaseURL                   string `json:"baseURL"`
	FrontendURL               string `json:"frontendURL"`
	JWTSecret                 string `json:"jwtSecret"`
	CaptchaEnabled            bool   `json:"captchaEnabled"`
	CaptchaVerifyUrl          string `json:"captchaVerifyUrl"`
	CaptchaSecret             string `json:"captchaSecret"`
	CaptchaSiteKey            string `json:"captchaSiteKey"`
	EmailEnabled              bool   `json:"emailEnabled"`
	EmailHost                 string `json:"emailHost"`
	EmailPort                 int    `json:"emailPort"`
	EmailUsername             string `json:"emailUsername"`
	EmailPassword             string `json:"emailPassword"`
	MinNameLength             int    `json:"minNameLength"`
	MaxNameLength             int    `json:"maxNameLength"`
	MinDescriptionLength      int    `json:"minDescriptionLength"`
	MaxDescriptionLength      int    `json:"maxDescriptionLength"`
	MinPasswordLength         int    `json:"minPasswordLength"`
	MaxPasswordLength         int    `json:"maxPasswordLength"`
	MinEmailLength            int    `json:"minEmailLength"`
	MaxEmailLength            int    `json:"maxEmailLength"`
	MaxProfilePictureFileSize int64  `json:"maxProfilePictureFileSize"`
	BcryptCost                int    `json:"bcryptCost"`
	PBKDF2Iterations          int    `json:"pbkdf2Iterations"`
	RecoveryCodeCount         uint   `json:"recoveryCodeCount"`
	LoginTokenLifetime        int64  `json:"loginTokenLifetime"`
	EmailCodeLifetime         int64  `json:"emailCodeLifetime"`
	AuthTokenLifetime         int64  `json:"authTokenLifetime"`
	RefreshTokenLifetime      int64  `json:"refreshTokenLifetime"`
	SendEmailTimeout          int64  `json:"sendEmailTimeout"`
	MaxPageSize               int    `json:"maxPageSize"`
	FrontendRoot              string `json:"frontendRoot"`
	IDProvider                string `json:"idProvider"`
	ClientID                  string `json:"clientID"`
	ClientSecret              string `json:"clientSecret"`
}

var defaultData = ConfigData{
	ServerPort:                80,
	DomainName:                "",
	BaseURL:                   "",
	FrontendURL:               "https://hbank.julianh.de",
	MinNameLength:             3,
	MaxNameLength:             30,
	MinDescriptionLength:      0,
	MaxDescriptionLength:      256,
	MinPasswordLength:         6,
	MinEmailLength:            3,
	MaxPasswordLength:         64,
	MaxEmailLength:            64,
	MaxProfilePictureFileSize: 10000000, // 10 MB
	BcryptCost:                10,
	PBKDF2Iterations:          10000,
	RecoveryCodeCount:         5,
	LoginTokenLifetime:        5 * 60,
	EmailCodeLifetime:         5 * 60,
	AuthTokenLifetime:         10 * 60,
	RefreshTokenLifetime:      1 * 365 * 24 * 60 * 60,
	SendEmailTimeout:          2 * 60,
	MaxPageSize:               100,
	IDProvider:                "https://id.julianh.de",
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

	if Data.SSL {
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

		if Data.ServerPort == 80 {
			Data.ServerPort = 443
		}
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
		Data.DomainName = "hbank"
	}

	if len(Data.JWTSecret) < 10 {
		log.Fatalln("ERROR: Please specify a jwt secret (>=10 characters)")
	}

	if Data.ClientID == "" {
		log.Fatalln("ERROR: Empty OAuth client ID")
	}
	if Data.ClientSecret == "" {
		log.Fatalln("ERROR: Empty OAuth client secret")
	}

	if Data.BaseURL == "" {
		log.Fatalln("ERROR: No base URL specified. Using default: https://hbank.julianh.de")
		Data.DomainName = "https://hbank.julianh.de"
	}

	if Data.FrontendURL == "" {
		log.Println("WARNING: Empty frontend URL. Using default: https://hbank.julianh.de")
		Data.FrontendURL = "https://hbank.julianh.de"
	}
}
