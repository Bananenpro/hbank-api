package handlers

import (
	"github.com/Bananenpro/oidc-client/oidc"

	"github.com/Bananenpro/hbank-api/models"
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
