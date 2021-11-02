package main

import (
	"context"
	"fmt"
	"log"
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
)

func main() {
	config.Load([]string{"config.json", xdg.ConfigHome + "/hbank/config.json"})
	services.LoadTranslations()

	services.EmailAuthenticate()

	r := router.New()

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

	v1 := r.Group("/v1")
	handler.RegisterV1(v1)

	go func() {
		err = r.StartTLS(fmt.Sprintf(":%d", config.Data.ServerPort), config.Data.SSLCertPath, config.Data.SSLKeyPath)
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
