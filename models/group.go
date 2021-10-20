package models

import (
	"github.com/google/uuid"
)

type GroupStore interface {
	GetAllByUser(user *User, page, pageSize int, descending bool) ([]Group, error)
	GetById(id uuid.UUID) (*Group, error)
	Create(group *Group) error
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

	GetTransactionLog(group *Group, user *User, page, pageSize int, oldestFirst bool) ([]TransactionLogEntry, error)
	GetTransactionLogEntryById(group *Group, id uuid.UUID) (*TransactionLogEntry, error)
	GetLastTransactionLogEntry(group *Group, user *User) (*TransactionLogEntry, error)
	GetUserBalance(group *Group, user *User) (int, error)
	CreateTransaction(group *Group, sender *User, receiver *User, title, description string, amount int) error
}

type Group struct {
	Base
	Name           string
	Description    string
	GroupPicture   []byte
	GroupPictureId uuid.UUID `gorm:"type:uuid"`

	Memberships []GroupMembership
}

type GroupMembership struct {
	Base
	GroupName string
	GroupId   uuid.UUID `gorm:"type:uuid"`
	UserName  string
	UserId    uuid.UUID `gorm:"type:uuid"`
	IsMember  bool
	IsAdmin   bool
}

type TransactionLogEntry struct {
	Base
	Title       string
	Description string
	Amount      int

	GroupId uuid.UUID `gorm:"type:uuid"`

	SenderId                uuid.UUID `gorm:"type:uuid"`
	NewBalanceSender        int
	BalanceDifferenceSender int

	ReceiverId                uuid.UUID `gorm:"type:uuid"`
	NewBalanceReceiver        int
	BalanceDifferenceReceiver int
}
