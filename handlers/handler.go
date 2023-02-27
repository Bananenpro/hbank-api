package handlers

import (
	"github.com/juho05/oidc-client/oidc"

	"github.com/juho05/hbank-api/models"
)

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
