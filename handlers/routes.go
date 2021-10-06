package handlers

import "github.com/labstack/echo/v4"

func (h *Handler) RegisterV1(v1 *echo.Group) {
	v1.GET("/status", h.Status)

	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.GET("/confirmEmail", h.SendConfirmEmail)
	auth.POST("/confirmEmail", h.VerifyConfirmEmailCode)
	auth.POST("/login", h.Login)

	twoFactor := auth.Group("/twoFactor")
	twoFactor.POST("/otp/activate", h.Activate2FAOTP)
	twoFactor.POST("/otp/verify", h.VerifyOTPCode)
}
