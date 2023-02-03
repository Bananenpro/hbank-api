package handlers

import (
	"github.com/labstack/echo/v4"

	"github.com/Bananenpro/hbank-api/router/middlewares"
)

func (h *Handler) RegisterAPI(api *echo.Group) {
	api.GET("/status", h.Status)

	jwt := middlewares.JWT(h.oidcClient, h.userStore)

	auth := api.Group("/auth")
	auth.GET("/login", h.Login)
	auth.GET("/callback", h.LoginCallback)
	auth.GET("/refresh", func(c echo.Context) error {
		return nil
	}, jwt)
	auth.POST("/logout", h.Logout)

	api.GET("/user", h.GetUsers, jwt)
	api.GET("/user/:id", h.GetUser, jwt)
	api.PUT("/user", h.UpdateUser, jwt)
	api.POST("/user/delete", h.DeleteUser, jwt)

	user := api.Group("/user")

	user.GET("/cash/current", h.GetCurrentCash, jwt)
	user.GET("/cash/:id", h.GetCashLogEntryById, jwt)
	user.GET("/cash", h.GetCashLog, jwt)
	user.POST("/cash", h.AddCashLogEntry, jwt)

	api.GET("/group", h.GetGroups, jwt)
	api.GET("/group/:id", h.GetGroupById, jwt)
	api.POST("/group", h.CreateGroup, jwt)
	api.PUT("/group/:id", h.UpdateGroup, jwt)

	group := api.Group("/group")
	group.GET("/:id/member", h.GetGroupMembers, jwt)
	group.DELETE("/:id/member", h.LeaveGroup, jwt)
	group.GET("/:id/admin", h.GetGroupAdmins, jwt)
	group.POST("/:id/admin", h.AddGroupAdmin, jwt)
	group.DELETE("/:id/admin", h.RemoveAdminRights, jwt)
	group.GET("/:id/user", h.GetGroupUsers, jwt)
	group.GET("/:id/picture", h.GetGroupPicture, jwt)
	group.POST("/:id/picture", h.SetGroupPicture, jwt)
	group.DELETE("/:id/picture", h.RemoveGroupPicture, jwt)

	group.GET("/:id/transaction/balance", h.GetBalance, jwt)
	group.GET("/:id/transaction/:transactionId", h.GetTransactionById, jwt)
	group.GET("/:id/transaction", h.GetTransactionLog, jwt)
	group.POST("/:id/transaction", h.CreateTransaction, jwt)

	group.GET("/:id/invitation", h.GetInvitationsByGroup, jwt)
	group.GET("/invitation", h.GetInvitationsByUser, jwt)
	group.GET("/invitation/:id", h.GetInvitationById, jwt)
	group.POST("/:id/invitation", h.CreateInvitation, jwt)
	group.POST("/invitation/:id", h.AcceptInvitation, jwt)
	group.DELETE("/invitation/:id", h.DenyInvitation, jwt)

	group.GET("/:id/paymentPlan/:paymentPlanId", h.GetPaymentPlanById, jwt)
	group.GET("/:id/paymentPlan", h.GetPaymentPlans, jwt)
	group.GET("/:id/paymentPlan/nextPayment", h.GetPaymentPlanNextPayments, jwt)
	group.POST("/:id/paymentPlan", h.CreatePaymentPlan, jwt)
	group.PUT("/:id/paymentPlan/:paymentPlanId", h.UpdatePaymentPlan, jwt)
	group.DELETE("/:id/paymentPlan/:paymentPlanId", h.DeletePaymentPlan, jwt)

	group.GET("/:id/total", h.GetTotalMoney, jwt)
}
