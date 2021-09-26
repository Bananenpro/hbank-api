package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/handlers"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	e := echo.New()

	db, err := gorm.Open(sqlite.Open("database.sqlite"))
	if err != nil {
		log.Fatalln("error opening database:", err)
	}
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(models.DBContextKey, db)
			return next(c)
		}
	})

	e.GET("/v1/health-check", handlers.HealthCheck)

	e.Logger.Fatal(e.Start(":8080"))
}
