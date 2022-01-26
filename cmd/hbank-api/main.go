package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/db"
	"github.com/Bananenpro/hbank-api/handlers"
	"github.com/Bananenpro/hbank-api/router"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/adrg/xdg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func serveFrontend(router *echo.Echo, path string) {
	if _, err := os.Stat(config.Data.FrontendRoot); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Fatalf("Couldn't find '%s'!", config.Data.FrontendRoot)
		} else {
			log.Fatalf("Couldn't open '%s': %s", config.Data.FrontendRoot, err)
		}
	}
	mime.AddExtensionType(".html", "text/html")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	router.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  config.Data.FrontendRoot,
		HTML5: true,
	}))
}

func main() {
	config.Load([]string{"config.json", xdg.ConfigHome + "/hbank/config.json"})
	services.LoadTranslations()

	services.EmailAuthenticate()

	r := router.New()

	if config.Data.FrontendRoot != "" {
		serveFrontend(r, config.Data.FrontendRoot)
	}

	database, err := db.NewSqlite("database.sqlite")
	if err != nil {
		log.Fatalln("Couldn't connect to database:", err)
	}
	err = db.AutoMigrate(database)
	if err != nil {
		log.Fatalln("Couldn't auto migrate database:", err)
	}

	us := db.NewUserStore(database)
	gs := db.NewGroupStore(database)
	handler := handlers.New(us, gs)

	api := r.Group("/api")
	handler.RegisterApi(api)

	go func() {
		if config.Data.SSL {
			err = r.StartTLS(fmt.Sprintf(":%d", config.Data.ServerPort), config.Data.SSLCertPath, config.Data.SSLKeyPath)
		} else {
			err = r.Start(fmt.Sprintf(":%d", config.Data.ServerPort))
		}
		if err != nil && err != http.ErrServerClosed {
			r.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.Shutdown(ctx); err != nil {
		r.Logger.Fatal(err)
	}
}
