package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/juho05/oidc-client/oidc"
	"github.com/labstack/echo/v4"

	hbank "github.com/juho05/h-bank"

	"github.com/juho05/h-bank/config"
	"github.com/juho05/h-bank/db"
	"github.com/juho05/h-bank/handlers"
	"github.com/juho05/h-bank/router"
	"github.com/juho05/h-bank/services"
)

func run(r *echo.Echo) error {
	hbank.Initialize()

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

	mux := http.NewServeMux()
	mux.Handle("/api/", r)
	mux.Handle("/", handlers.NewFrontendHandler())

	handler := handlers.New(us, gs, oidcClient)

	api := r.Group("/api")
	handler.RegisterAPI(api)

	go func() {
		if config.Data.SSL {
			err = http.ListenAndServeTLS(fmt.Sprintf(":%d", config.Data.ServerPort), config.Data.SSLCertPath, config.Data.SSLKeyPath, mux)
		} else {
			err = http.ListenAndServe(fmt.Sprintf(":%d", config.Data.ServerPort), mux)
		}
		if err != nil && err != http.ErrServerClosed {
			r.Logger.Error(err)
		}
	}()

	log.Printf("Listening on port %d", config.Data.ServerPort)

	StartPaymentPlanTicker(us, gs)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	close(StopPaymentPlanTicker)
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
