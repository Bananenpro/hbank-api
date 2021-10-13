package router

import (
	"github.com/Bananenpro/hbank-api/router/middlewares"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func New() *echo.Echo {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())

	e.Use(middleware.Recover())
	e.Use(middleware.Secure())

	e.Use(middlewares.Lang)

	return e
}
