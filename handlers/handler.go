package handlers

import (
	"mime"

	"github.com/juho05/oidc-client/oidc"

	"github.com/juho05/h-bank/models"
)

func init() {
	mime.AddExtensionType(".html", "text/html")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
}

type Handler struct {
	userStore  models.UserStore
	groupStore models.GroupStore
	oidcClient *oidc.Client
}

func New(userStore models.UserStore, groupStore models.GroupStore, oidcClient *oidc.Client) *Handler {
	return &Handler{
		userStore:  userStore,
		groupStore: groupStore,
		oidcClient: oidcClient,
	}
}
