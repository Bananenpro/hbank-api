package hbank

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/juho05/hbank-api/config"
)

//go:embed all:frontend/dist
var frontendFS embed.FS
var FrontendFS fs.FS

var (
	StartTime          = time.Now()
	DevFrontendEnabled = false
)

func Initialize() {
	var err error

	if config.Data.DevFrontend != "" {
		_, err = http.Get(config.Data.DevFrontend)
		DevFrontendEnabled = err == nil
		if DevFrontendEnabled {
			log.Println("Forwarding frontend requests to", config.Data.DevFrontend)
			return
		} else {
			log.Println("WARNING: Dev frontend at %s is not reachable", config.Data.DevFrontend)
		}
	}

	if config.Data.FrontendDir != "" {
		FrontendFS = os.DirFS(config.Data.FrontendDir)
		log.Println("Using custom frontend directory:", config.Data.FrontendDir)
		return
	}

	FrontendFS, err = fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
}
