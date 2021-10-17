package models

import (
	"github.com/google/uuid"
)

type GroupStore interface {
	GetAllByUser(user *User, page, pageSize int, descending bool) ([]Group, error)
	GetById(id uuid.UUID) (*Group, error)
	Create(user *User, group *Group) error
	Update(group *Group) error
	Delete(group *Group) error
	DeleteById(id uuid.UUID) error

	GetGroupPicture(group *Group) ([]byte, error)

	GetMembers(group *Group, page, pageSize int, descending bool) ([]User, error)
	IsMember(group *Group, user *User) (bool, error)
	AddMember(group *Group, user *User) error
	RemoveMember(group *Group, user *User) error
}

type Group struct {
	Base
	Name           string
	Description    string
	GroupPicture   []byte
	GroupPictureId uuid.UUID `gorm:"type:uuid"`

	Members []User `gorm:"many2many:group_members"`
}