package handlers

import (
	"github.com/Bananenpro/hbank-api/router/middlewares"
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterV1(v1 *echo.Group) {
	v1.GET("/status", h.Status)

	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.GET("/confirmEmail/:email", h.SendConfirmEmail)
	auth.POST("/confirmEmail", h.VerifyConfirmEmailCode)
	auth.POST("/passwordAuth", h.PasswordAuth)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout, middlewares.JWT)
	auth.POST("/changePassword", h.ChangePassword, middlewares.JWT)
	auth.POST("/forgotPassword", h.ForgotPassword)
	auth.POST("/resetPassword", h.ResetPassword)
	auth.POST("/requestChangeEmail", h.RequestChangeEmail, middlewares.JWT)
	auth.POST("/changeEmail", h.ChangeEmail, middlewares.JWT)

	twoFactor := auth.Group("/twoFactor")
	twoFactor.POST("/otp/activate", h.Activate2FAOTP)
	twoFactor.POST("/otp/get", h.GetOTPQRCode, middlewares.JWT)
	twoFactor.POST("/otp/verify", h.VerifyOTPCode)
	twoFactor.POST("/otp/new", h.NewOTP, middlewares.JWT)

	twoFactor.POST("/recovery/verify", h.VerifyRecoveryCode)
	twoFactor.POST("/recovery/new", h.NewRecoveryCodes, middlewares.JWT)

	v1.GET("/user", h.GetUsers, middlewares.JWT)
	v1.GET("/user/:id", h.GetUser, middlewares.JWT)
}
