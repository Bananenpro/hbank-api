package models

import (
	"github.com/juho05/h-bank/services"
)

type GroupStore interface {
	GetAllByUser(user *User, page, pageSize int, descending bool) ([]Group, error)
	Count(user *User) (int64, error)
	GetById(id string) (*Group, error)
	Create(group *Group) error
	Update(group *Group) error
	Delete(group *Group) error
	DeleteById(id string) error

	GetGroupPicture(group *Group, size services.PictureSize) ([]byte, error)
	UpdateGroupPicture(group *Group, pic *GroupPicture) error

	GetMembers(except *User, searchInput string, group *Group, page, pageSize int, descending bool) ([]User, error)
	MemberCount(group *Group) (int64, error)
	IsMember(group *Group, user *User) (bool, error)
	AddMember(group *Group, user *User) error
	RemoveMember(group *Group, user *User) error

	GetAdmins(except *User, searchInput string, group *Group, page, pageSize int, descending bool) ([]User, error)
	AdminCount(group *Group) (int64, error)
	IsAdmin(group *Group, user *User) (bool, error)
	AddAdmin(group *Group, user *User) error
	RemoveAdmin(group *Group, user *User) error

	GetMemberships(except *User, searchInput string, group *Group, page, pageSize int, descending bool) ([]GroupMembership, error)
	MembershipCount(group *Group) (int64, error)

	IsInGroup(group *Group, user *User) (bool, error)
	GetUserCount(group *Group) (int64, error)

	GetTransactionLog(group *Group, user *User, searchInput string, page, pageSize int, oldestFirst bool) ([]TransactionLogEntry, error)
	TransactionLogEntryCount(group *Group, user *User) (int64, error)
	GetBankTransactionLog(group *Group, searchInput string, page, pageSize int, oldestFirst bool) ([]TransactionLogEntry, error)
	BankTransactionLogEntryCount(group *Group) (int64, error)
	GetTransactionLogEntryById(group *Group, id string) (*TransactionLogEntry, error)
	GetLastTransactionLogEntry(group *Group, user *User) (*TransactionLogEntry, error)
	GetUserBalance(group *Group, user *User) (int, error)
	CreateTransaction(group *Group, senderIsBank, receiverIsBank bool, sender *User, receiver *User, title, description string, amount int) (*TransactionLogEntry, error)
	CreateTransactionFromPaymentPlan(group *Group, senderIsBank, receiverIsBank bool, sender *User, receiver *User, title, description string, amount int, paymentPlanId string) (*TransactionLogEntry, error)

	CreateInvitation(group *Group, user *User, message string) (*GroupInvitation, error)
	GetInvitationById(id string) (*GroupInvitation, error)
	GetInvitationsByGroup(group *Group, page, pageSize int, oldestFirst bool) ([]GroupInvitation, error)
	InvitationCountByGroup(group *Group) (int64, error)
	GetInvitationsByUser(user *User, page, pageSize int, oldestFirst bool) ([]GroupInvitation, error)
	InvitationCountByUser(user *User) (int64, error)
	GetInvitationByGroupAndUser(group *Group, user *User) (*GroupInvitation, error)
	DeleteInvitation(invitation *GroupInvitation) error

	GetPaymentPlans(group *Group, user *User, searchInput string, page, pageSize int, descending bool) ([]PaymentPlan, error)
	PaymentPlanCount(group *Group, user *User) (int64, error)
	GetBankPaymentPlans(group *Group, searchInput string, page, pageSize int, descending bool) ([]PaymentPlan, error)
	BankPaymentPlanCount(group *Group) (int64, error)
	GetPaymentPlansThatNeedToBeExecuted() ([]PaymentPlan, error)
	GetPaymentPlanById(group *Group, id string) (*PaymentPlan, error)
	CreatePaymentPlan(group *Group, senderIsBank, receiverIsBank bool, sender *User, receiver *User, name, description string, amount, repeats, schedule int, scheduleUnit string, firstPayment int64) (*PaymentPlan, error)
	UpdatePaymentPlan(paymentPlan *PaymentPlan) error
	DeletePaymentPlan(paymentPlan *PaymentPlan) error

	GetTotalMoney(group *Group) (int, error)

	AreInSameGroup(userId1, userId2 string) (bool, error)
}

type Group struct {
	Base
	Name           string
	Description    string
	GroupPicture   *GroupPicture `gorm:"constraint:OnDelete:CASCADE"`
	GroupPictureId string

	Memberships []GroupMembership
	Invitations []GroupInvitation
}

type GroupPicture struct {
	Base

	Tiny   []byte
	Small  []byte
	Medium []byte
	Large  []byte
	Huge   []byte

	GroupId string
}

type GroupMembership struct {
	Base
	GroupId   string
	GroupName string
	UserId    string
	UserName  string
	IsMember  bool
	IsAdmin   bool
}

type GroupInvitation struct {
	Base
	GroupName string
	Message   string
	GroupId   string
	UserId    string
}

type TransactionLogEntry struct {
	Base
	Title       string
	Description string
	Amount      int

	GroupId string

	SenderIsBank            bool
	SenderId                string
	NewBalanceSender        int
	BalanceDifferenceSender int

	ReceiverIsBank            bool
	ReceiverId                string
	NewBalanceReceiver        int
	BalanceDifferenceReceiver int

	PaymentPlanId string
}

const (
	ScheduleUnitDay   = "day"
	ScheduleUnitWeek  = "week"
	ScheduleUnitMonth = "month"
	ScheduleUnitYear  = "year"
)

type PaymentPlan struct {
	Base
	Name        string
	Description string

	Amount int

	// negative payment count for unlimited payments
	PaymentCount int

	NextExecute  int64
	Schedule     int
	ScheduleUnit string

	SenderIsBank bool
	SenderId     string

	ReceiverIsBank bool
	ReceiverId     string

	GroupId string
}
