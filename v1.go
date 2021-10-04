package main

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/Bananenpro05/hbank2-api/handlers"
)

func registerV1Routes(e *echo.Echo) {
	v1 := e.Group("/v1")
	v1StatusRoute(v1)

	auth := v1.Group("/auth")
	v1AuthRoutes(auth)
}

func v1StatusRoute(e *echo.Group) {
	e.GET("/status", handlers.Status)
}

func v1AuthRoutes(e *echo.Group) {
	e.POST("/register", handlers.Register)
	e.GET("/confirmEmail", handlers.SendConfirmEmail)
	e.POST("/confirmEmail", handlers.VerifyConfirmEmailCode)
	e.POST("/login", handlers.Login)

	twoFactor := e.Group("/twoFactor")
	v1AuthTwoFactorRoutes(twoFactor)
}

func v1AuthTwoFactorRoutes(e *echo.Group) {
	e.POST("/otp/activate", handlers.Activate2FAOTP)
	e.POST("/otp/verify", handlers.VerifyOTPCode)
}
