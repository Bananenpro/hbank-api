package handlers

import "github.com/Bananenpro/hbank-api/models"

type Handler struct {
	userStore  models.UserStore
	groupStore models.GroupStore
}

func New(userStore models.UserStore, groupStore models.GroupStore) *Handler {
	return &Handler{
		userStore:  userStore,
		groupStore: groupStore,
	}
}
