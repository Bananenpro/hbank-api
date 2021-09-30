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
}

var defaultData = ConfigData{
	ServerPort:     8080,
	CaptchaEnabled: false,
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
		log.Println("Invalid port number. Using default port: ", defaultData.ServerPort)
		Data.ServerPort = defaultData.ServerPort
	}

	if Data.CaptchaEnabled {
		if Data.HCaptchaSecret == "" {
			log.Println("No captcha secret specified. Disabling captcha.")
			Data.CaptchaEnabled = false
		}
		if Data.HCaptchaSiteKey == "" {
			log.Println("No captcha site key specified. Disabling captcha.")
			Data.CaptchaEnabled = false
		}
	}
}
