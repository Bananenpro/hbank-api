package main

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/handlers"
)

func registerV1Routes(e *echo.Echo) {
	v1 := e.Group("/v1")
	v1StatusRoute(v1)
}

func v1StatusRoute(e *echo.Group) {
	e.GET("/status", handlers.Status)
}
