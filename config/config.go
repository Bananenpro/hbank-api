package config

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
)

type DBEngine string

const (
	DBSqlite   DBEngine = "sqlite"
	DBPostgres DBEngine = "postgres"
)

type ConfigData struct {
	Debug                     bool     `json:"debug"`
	DBEngine                  DBEngine `json:"dbEngine"`
	DBPath                    string   `json:"dbPath"`
	DBHost                    string   `json:"dbHost"`
	DBPort                    int      `json:"dbPort"`
	DBUser                    string   `json:"dbUser"`
	DBPassword                string   `json:"dbPassword"`
	DBName                    string   `json:"dbName"`
	DBVerbose                 bool     `json:"dbVerbose"`
	ServerPort                int      `json:"serverPort"`
	SSL                       bool     `json:"ssl"`
	SSLCertPath               string   `json:"sslCertPath"`
	SSLKeyPath                string   `json:"sslKeyPath"`
	BaseURL                   string   `json:"baseURL"`
	DomainName                string   `json:"-"`
	EmailEnabled              bool     `json:"emailEnabled"`
	EmailHost                 string   `json:"emailHost"`
	EmailPort                 int      `json:"emailPort"`
	EmailUsername             string   `json:"emailUsername"`
	EmailPassword             string   `json:"emailPassword"`
	MinNameLength             int      `json:"minNameLength"`
	MaxNameLength             int      `json:"maxNameLength"`
	MinDescriptionLength      int      `json:"minDescriptionLength"`
	MaxDescriptionLength      int      `json:"maxDescriptionLength"`
	MaxProfilePictureFileSize int64    `json:"maxProfilePictureFileSize"`
	MaxPageSize               int      `json:"maxPageSize"`
	IDProvider                string   `json:"idProvider"`
	InternalIDProvider        string `json:"internalIDProvider"`
	ClientID                  string   `json:"clientID"`
	ClientSecret              string   `json:"clientSecret"`
	DevFrontend               string   `json:"devFrontend"`
	FrontendDir               string   `json:"frontendDir"`
}

var defaultData = ConfigData{
	ServerPort:                80,
	BaseURL:                   "",
	DBPath:                    "database.sqlite",
	MinNameLength:             3,
	MaxNameLength:             30,
	MinDescriptionLength:      0,
	MaxDescriptionLength:      256,
	MaxProfilePictureFileSize: 10000000, // 10 MB
	MaxPageSize:               100,
	IDProvider:                "",
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
	if Data.ServerPort <= 0 || Data.ServerPort > 65353 {
		if Data.ServerPort != 0 {
			log.Println("WARNING: Invalid port number. Using default port: ", defaultData.ServerPort)
		}
		Data.ServerPort = defaultData.ServerPort
	}

	switch Data.DBEngine {
	case DBSqlite:
		if Data.DBPath == "" {
			log.Fatalln("ERROR: dbPath is required when using SQLite as DB engine")
		}
	case DBPostgres:
		if Data.DBHost == "" {
			log.Fatalln("ERROR: dbHost is required when using Postgres as DB engine")
		}
		if Data.DBPort == 0 {
			log.Fatalln("ERROR: dbPort is required when using Postgres as DB engine")
		}
		if Data.DBUser == "" {
			log.Fatalln("ERROR: dbUser is required when using Postgres as DB engine")
		}
		if Data.DBPassword == "" {
			log.Fatalln("ERROR: dbPassword is required when using Postgres as DB engine")
		}
		if Data.DBName == "" {
			log.Fatalln("ERROR: dbName is required when using Postgres as DB engine")
		}
	case "":
		log.Fatalln("ERROR: dbEngine config option is required.")
	default:
		log.Fatalf("ERROR: Invalid dbEngine value. Supported values: %s, %s", DBSqlite, DBPostgres)
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

	if Data.ClientID == "" {
		log.Fatalln("ERROR: Empty OAuth client ID")
	}
	if Data.ClientSecret == "" {
		log.Fatalln("ERROR: Empty OAuth client secret")
	}

	if Data.BaseURL == "" {
		log.Fatalln("ERROR: No base URL specified")
	}

	if strings.TrimSpace(Data.DomainName) == "" {
		baseURL, err := url.Parse(Data.BaseURL)
		if err != nil {
			log.Fatalln("ERROR: Invalid base URL")
		}
		Data.DomainName = baseURL.Hostname()
	}

	if Data.IDProvider == "" {
		log.Fatalln("ERROR: No ID provider specified")
	}
	if Data.InternalIDProvider == "" {
		Data.InternalIDProvider = Data.IDProvider
	}

	if _, err := url.Parse(Data.DevFrontend); err != nil {
		log.Println("WARNING: Invalid dev frontend URL:", err)
		Data.DevFrontend = ""
	}
}
