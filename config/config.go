package config

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigData struct {
	ServerPort      int    `json:"server_port"`
	CaptchaEnabled  bool   `json:"captcha_enabled"`
	HCaptchaSecret  string `json:"h_captcha_secret"`
	HCaptchaSiteKey string `json:"h_captcha_site_key"`
	EmailEnabled    bool   `json:"email_enabled"`
	EmailHost       string `json:"email_host"`
	EmailPort       int    `json:"email_port"`
	EmailUsername   string `json:"email_username"`
	EmailPassword   string `json:"email_password"`
}

var defaultData = ConfigData{
	ServerPort: 8080,
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
		if Data.HCaptchaSecret == "" {
			log.Println("WARNING: No captcha secret specified. Disabling captcha.")
			Data.CaptchaEnabled = false
		}
		if Data.HCaptchaSiteKey == "" {
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
}
