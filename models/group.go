package models

import (
	"github.com/google/uuid"
)

type GroupStore interface {
	GetAllByMember(member *User, page, pageSize int, descending bool) ([]Group, error)
	GetAllByAdmin(admin *User, page, pageSize int, descending bool) ([]Group, error)
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

	GetAdmins(group *Group, page, pageSize int, descending bool) ([]User, error)
	IsAdmin(group *Group, user *User) (bool, error)
	AddAdmin(group *Group, user *User) error
	RemoveAdmin(group *Group, user *User) error

	IsInGroup(group *Group, user *User) (bool, error)
}

type Group struct {
	Base
	Name           string
	Description    string
	GroupPicture   []byte
	GroupPictureId uuid.UUID `gorm:"type:uuid"`

	Members []User `gorm:"many2many:group_members"`
	Admins  []User `gorm:"many2many:group_admins"`
}
