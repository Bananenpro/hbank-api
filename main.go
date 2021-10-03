package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/adrg/xdg"
	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/config"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"gitlab.com/Bananenpro05/hbank2-api/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	config.Load([]string{"config.json", xdg.ConfigHome + "/hbank/config.json"})

	services.EmailAuthenticate()

	e := echo.New()

	db, err := gorm.Open(sqlite.Open("database.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalln("error opening database:", err)
	}

	errs := models.AutoMigrate(db)
	for _, err := range errs {
		if err != nil {
			log.Println("error migrating database:", err)
		}
	}
	if len(errs) > 0 {
		os.Exit(1)
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(models.DBContextKey, db)
			return next(c)
		}
	})

	registerV1Routes(e)

	go func() {
		if err := e.Start(fmt.Sprintf(":%d", config.Data.ServerPort)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
