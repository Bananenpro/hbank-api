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
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/juho05/oidc-client/oidc"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/juho05/hbank-api/config"
	"github.com/juho05/hbank-api/db"
	"github.com/juho05/hbank-api/handlers"
	"github.com/juho05/hbank-api/router"
	"github.com/juho05/hbank-api/services"
)

func serveFrontend(router *echo.Echo) {
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

func run(r *echo.Echo) error {
	if config.Data.FrontendRoot != "" {
		serveFrontend(r)
	}

	database, err := db.NewSqlite("database.sqlite?_pragma=foreign_keys(1)&_pragma=busy_timeout(3000)&_pragma=journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("Couldn't connect to database: %w", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("Failed to get generic SQL interface: %w", err)
	}
	defer sqlDB.Close()
	err = db.AutoMigrate(database)
	if err != nil {
		return fmt.Errorf("Couldn't auto migrate database: %w", err)
	}

	us := db.NewUserStore(database)
	gs := db.NewGroupStore(database)

	oidcClient, err := oidc.NewClient(config.Data.IDProvider, oidc.ClientConfig{
		ClientID:     config.Data.ClientID,
		ClientSecret: config.Data.ClientSecret,
		RedirectURI:  config.Data.BaseURL + "/api/auth/callback",
	})
	if err != nil {
		return fmt.Errorf("Couldn't create OIDC client: %w", err)
	}

	handler := handlers.New(us, gs, oidcClient)

	api := r.Group("/api")
	handler.RegisterAPI(api)

	go func() {
		if config.Data.SSL {
			err = r.StartTLS(fmt.Sprintf(":%d", config.Data.ServerPort), config.Data.SSLCertPath, config.Data.SSLKeyPath)
		} else {
			err = r.Start(fmt.Sprintf(":%d", config.Data.ServerPort))
		}
		if err != nil && err != http.ErrServerClosed {
			r.Logger.Error(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func main() {
	config.Load([]string{"config.json", xdg.ConfigHome + "/hbank/config.json"})
	services.LoadTranslations()

	services.EmailAuthenticate()

	r := router.New()
	if err := run(r); err != nil {
		r.Logger.Fatal(err)
	}
}
