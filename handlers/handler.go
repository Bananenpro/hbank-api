package handlers

import "github.com/Bananenpro/hbank2-api/models"

type Handler struct {
	userStore models.UserStore
}

func New(userStore models.UserStore) *Handler {
	return &Handler{
		userStore: userStore,
	}
}
