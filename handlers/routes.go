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
	v1.PUT("/user", h.UpdateUser, middlewares.JWT)
	v1.DELETE("/user/:id", h.DeleteUserByConfirmEmailCode)
	v1.DELETE("/user", h.DeleteUser, middlewares.JWT)
	v1.POST("/user/picture", h.SetProfilePicture, middlewares.JWT)
	v1.GET("/user/:id/picture", h.GetProfilePicture, middlewares.JWT)

	user := v1.Group("/user")
	user.GET("/cash/current", h.GetCurrentCash, middlewares.JWT)
	user.GET("/cash/:id", h.GetCashLogEntryById, middlewares.JWT)
	user.GET("/cash", h.GetCashLog, middlewares.JWT)
	user.POST("/cash", h.AddCashLogEntry, middlewares.JWT)

	v1.GET("/group", h.GetGroups, middlewares.JWT)
	v1.GET("/group/:id", h.GetGroupById, middlewares.JWT)
	v1.POST("/group", h.CreateGroup, middlewares.JWT)

	group := v1.Group("/group")
	group.GET("/:id/member", h.GetGroupMembers, middlewares.JWT)
	group.GET("/:id/admin", h.GetGroupAdmins, middlewares.JWT)
	group.GET("/:id/picture", h.GetGroupPicture, middlewares.JWT)
	group.POST("/:id/picture", h.SetGroupPicture, middlewares.JWT)

	group.GET("/:id/transaction/balance", h.GetBalance, middlewares.JWT)
	group.GET("/:id/transaction/:transactionId", h.GetTransactionById, middlewares.JWT)
	group.GET("/:id/transaction", h.GetTransactionLog, middlewares.JWT)
	group.POST("/:id/transaction", h.CreateTransaction, middlewares.JWT)
}
