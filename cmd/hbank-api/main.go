package main

import (
	"fmt"
	"log"

	"github.com/Bananenpro/hbank2-api/config"
	"github.com/Bananenpro/hbank2-api/db"
	"github.com/Bananenpro/hbank2-api/handlers"
	"github.com/Bananenpro/hbank2-api/router"
	"github.com/Bananenpro/hbank2-api/services"
	"github.com/adrg/xdg"
)

func main() {
	config.Load([]string{"config.json", xdg.ConfigHome + "/hbank/config.json"})

	services.EmailAuthenticate()

	r := router.New()

	database, err := db.NewSqlite("database.sqlite")
	if err != nil {
		log.Fatalln("Couldn't connect to database")
	}
	err = db.AutoMigrate(database)
	if err != nil {
		log.Fatalln("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)
	handler := handlers.New(us)

	v1 := r.Group("/v1")
	handler.RegisterV1(v1)

	r.Logger.Fatal(r.StartTLS(fmt.Sprintf(":%d", config.Data.ServerPort), config.Data.SSLCertPath, config.Data.SSLKeyPath))
}
