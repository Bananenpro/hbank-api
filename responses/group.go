package responses

import (
	"bytes"

	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
)

type Balance struct {
	Base
	Balance int `json:"balance"`
}

type DeleteFailedBecauseOfSoleGroupAdmin struct {
	Base
	GroupIds []string `json:"groupIds"`
}

type PaymentPlanExecutionTimes struct {
	Base
	ExecutionTimes []int64 `json:"executionTimes"`
}

type group struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	GroupPictureId string `json:"groupPictureId"`
}

type groupDetailed struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	GroupPictureId string `json:"groupPictureId"`
	Member         bool   `json:"member"`
	Admin          bool   `json:"admin"`
}

type transaction struct {
	Id          string `json:"id"`
	Time        int64  `json:"time"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`

	GroupId string `json:"groupId"`

	Amount     int `json:"amount"`
	NewBalance int `json:"newBalance"`

	SenderId   string `json:"senderId"`
	ReceiverId string `json:"receiverId"`

	PaymentPlanId string `json:"paymentPlanId,omitempty"`
}

type bankTransaction struct {
	Id          string `json:"id"`
	Time        int64  `json:"time"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Amount      int    `json:"amount"`

	GroupId string `json:"groupId"`

	SenderId   string `json:"senderId"`
	ReceiverId string `json:"receiverId"`

	PaymentPlanId string `json:"paymentPlanId,omitempty"`
}

type paymentPlan struct {
	Id string `json:"id"`

	NextExecute int64 `json:"nextExecute"`

	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Schedule     int    `json:"schedule"`
	ScheduleUnit string `json:"scheduleUnit"`

	GroupId string `json:"groupId"`

	Amount int `json:"amount"`

	SenderId   string `json:"senderId,omitempty"`
	ReceiverId string `json:"receiverId,omitempty"`
}

type invitation struct {
	Id                string `json:"id"`
	Created           int64  `json:"created"`
	InvitationMessage string `json:"invitationMessage"`
	GroupName         string `json:"groupName,omitempty"`
	GroupId           string `json:"groupId,omitempty"`
	UserId            string `json:"userId,omitempty"`
}

func NewInvitations(invitations []models.GroupInvitation, count int64) interface{} {
	dtos := make([]invitation, len(invitations))
	for i, in := range invitations {
		dtos[i].Id = in.Id.String()
		dtos[i].Created = in.Created
		dtos[i].InvitationMessage = in.Message
		dtos[i].UserId = in.UserId.String()
		dtos[i].GroupName = in.GroupName
		dtos[i].GroupId = in.GroupId.String()
	}

	type invitationsResp struct {
		Base
		Count       int64        `json:"count"`
		Invitations []invitation `json:"invitations"`
	}

	return invitationsResp{
		Base: Base{
			Success: true,
		},
		Count:       count,
		Invitations: dtos,
	}
}

func NewInvitation(invitationModel *models.GroupInvitation) interface{} {
	type invitationResp struct {
		Base
		invitation
	}

	return invitationResp{
		Base: Base{
			Success: true,
		},
		invitation: invitation{
			Id:                invitationModel.Id.String(),
			Created:           invitationModel.Created,
			InvitationMessage: invitationModel.Message,
			GroupName:         invitationModel.GroupName,
			GroupId:           invitationModel.GroupId.String(),
			UserId:            invitationModel.UserId.String(),
		},
	}
}

func NewGroups(groups []models.Group, count int64) interface{} {
	groupDTOs := make([]group, len(groups))
	for i, g := range groups {
		groupDTOs[i].Id = g.Id.String()
		groupDTOs[i].Name = g.Name
		groupDTOs[i].Description = g.Description
		groupDTOs[i].GroupPictureId = g.GroupPictureId.String()
	}

	type groupsResp struct {
		Base
		Count  int64   `json:"count"`
		Groups []group `json:"groups"`
	}

	return groupsResp{
		Base: Base{
			Success: true,
		},
		Count:  count,
		Groups: groupDTOs,
	}
}

func NewGroup(group *models.Group, isMember, isAdmin bool) interface{} {
	type groupResp struct {
		Base
		groupDetailed
	}

	return groupResp{
		Base: Base{
			Success: true,
		},
		groupDetailed: groupDetailed{
			Id:             group.Id.String(),
			Name:           group.Name,
			Description:    group.Description,
			GroupPictureId: group.GroupPictureId.String(),
			Member:         isMember,
			Admin:          isAdmin,
		},
	}
}

func NewTransaction(transactionModel *models.TransactionLogEntry, user *models.User) interface{} {
	type transactionResp struct {
		Base
		transaction
	}

	isSender := bytes.Equal(user.Id[:], transactionModel.SenderId[:])

	newBalance := transactionModel.NewBalanceReceiver
	if isSender {
		newBalance = transactionModel.NewBalanceSender
	}

	transactionDTO := transaction{
		Id:          transactionModel.Id.String(),
		Time:        transactionModel.Created,
		Title:       transactionModel.Title,
		Description: transactionModel.Description,
		Amount:      transactionModel.Amount,
		NewBalance:  newBalance,
		GroupId:     transactionModel.GroupId.String(),
	}

	if transactionModel.ReceiverIsBank {
		transactionDTO.ReceiverId = "bank"
	} else {
		transactionDTO.ReceiverId = transactionModel.ReceiverId.String()
	}

	if transactionModel.SenderIsBank {
		transactionDTO.SenderId = "bank"
	} else {
		transactionDTO.SenderId = transactionModel.SenderId.String()
	}

	emptyId := uuid.UUID{}
	if !bytes.Equal(transactionModel.PaymentPlanId[:], emptyId[:]) {
		transactionDTO.PaymentPlanId = transactionModel.PaymentPlanId.String()
	}

	return transactionResp{
		Base: Base{
			Success: true,
		},
		transaction: transactionDTO,
	}
}

func NewBankTransaction(transactionModel *models.TransactionLogEntry) interface{} {
	type transactionResp struct {
		Base
		bankTransaction
	}

	transactionDTO := bankTransaction{
		Id:          transactionModel.Id.String(),
		Time:        transactionModel.Created,
		Title:       transactionModel.Title,
		Description: transactionModel.Description,
		Amount:      transactionModel.Amount,
		GroupId:     transactionModel.GroupId.String(),
	}

	if transactionModel.ReceiverIsBank {
		transactionDTO.ReceiverId = "bank"
	} else {
		transactionDTO.ReceiverId = transactionModel.ReceiverId.String()
	}

	if transactionModel.SenderIsBank {
		transactionDTO.SenderId = "bank"
	} else {
		transactionDTO.SenderId = transactionModel.SenderId.String()
	}

	emptyId := uuid.UUID{}
	if !bytes.Equal(transactionModel.PaymentPlanId[:], emptyId[:]) {
		transactionDTO.PaymentPlanId = transactionModel.PaymentPlanId.String()
	}

	return transactionResp{
		Base: Base{
			Success: true,
		},
		bankTransaction: transactionDTO,
	}
}

func NewTransactionLog(log []models.TransactionLogEntry, user *models.User, count int64) interface{} {
	type transactionsResp struct {
		Base
		Count        int64         `json:"count"`
		Transactions []transaction `json:"transactions"`
	}

	transactionDTOs := make([]transaction, len(log))

	for i, entry := range log {
		isSender := bytes.Equal(user.Id[:], entry.SenderId[:])

		newBalance := entry.NewBalanceReceiver
		if isSender {
			newBalance = entry.NewBalanceSender
		}

		transactionDTO := transaction{
			Id:         entry.Id.String(),
			Time:       entry.Created,
			Title:      entry.Title,
			Amount:     entry.Amount,
			NewBalance: newBalance,
			GroupId:    entry.GroupId.String(),
		}

		if entry.ReceiverIsBank {
			transactionDTO.ReceiverId = "bank"
		} else {
			transactionDTO.ReceiverId = entry.ReceiverId.String()
		}

		if entry.SenderIsBank {
			transactionDTO.SenderId = "bank"
		} else {
			transactionDTO.SenderId = entry.SenderId.String()
		}

		emptyId := uuid.UUID{}
		if !bytes.Equal(entry.PaymentPlanId[:], emptyId[:]) {
			transactionDTO.PaymentPlanId = entry.PaymentPlanId.String()
		}

		transactionDTOs[i] = transactionDTO
	}

	return transactionsResp{
		Base: Base{
			Success: true,
		},
		Count:        count,
		Transactions: transactionDTOs,
	}
}

func NewBankTransactionLog(log []models.TransactionLogEntry, count int64) interface{} {
	type transactionsResp struct {
		Base
		Count        int64             `json:"count"`
		Transactions []bankTransaction `json:"transactions"`
	}

	transactionDTOs := make([]bankTransaction, len(log))

	for i, entry := range log {
		transactionDTO := bankTransaction{
			Id:      entry.Id.String(),
			Time:    entry.Created,
			Title:   entry.Title,
			Amount:  entry.Amount,
			GroupId: entry.GroupId.String(),
		}

		if entry.ReceiverIsBank {
			transactionDTO.ReceiverId = "bank"
		} else {
			transactionDTO.ReceiverId = entry.ReceiverId.String()
		}

		if entry.SenderIsBank {
			transactionDTO.SenderId = "bank"
		} else {
			transactionDTO.SenderId = entry.SenderId.String()
		}

		emptyId := uuid.UUID{}
		if !bytes.Equal(entry.PaymentPlanId[:], emptyId[:]) {
			transactionDTO.PaymentPlanId = entry.PaymentPlanId.String()
		}

		transactionDTOs[i] = transactionDTO
	}

	return transactionsResp{
		Base: Base{
			Success: true,
		},
		Count:        count,
		Transactions: transactionDTOs,
	}
}

func NewDeleteFailedBecauseOfSoleGroupAdmin(groupIds []uuid.UUID, lang string) interface{} {
	ids := make([]string, len(groupIds))
	for i := range groupIds {
		ids[i] = groupIds[i].String()
	}

	return &DeleteFailedBecauseOfSoleGroupAdmin{
		Base: Base{
			Success: false,
			Message: services.Tr("Failed to delete user because he is the only admin of one or more groups", lang),
		},
		GroupIds: ids,
	}
}

func NewPaymentPlan(paymentPlanModel *models.PaymentPlan) interface{} {
	type paymentPlanResp struct {
		Base
		paymentPlan
	}

	paymentPlanDTO := paymentPlan{
		Id:           paymentPlanModel.Id.String(),
		NextExecute:  paymentPlanModel.NextExecute,
		Name:         paymentPlanModel.Name,
		Description:  paymentPlanModel.Description,
		Schedule:     paymentPlanModel.Schedule,
		ScheduleUnit: paymentPlanModel.ScheduleUnit,
		Amount:       paymentPlanModel.Amount,
		GroupId:      paymentPlanModel.GroupId.String(),
	}

	if paymentPlanModel.ReceiverIsBank {
		paymentPlanDTO.ReceiverId = "bank"
	} else {
		paymentPlanDTO.ReceiverId = paymentPlanModel.ReceiverId.String()
	}

	if paymentPlanModel.SenderIsBank {
		paymentPlanDTO.SenderId = "bank"
	} else {
		paymentPlanDTO.SenderId = paymentPlanModel.SenderId.String()
	}

	return paymentPlanResp{
		Base: Base{
			Success: true,
		},
		paymentPlan: paymentPlanDTO,
	}
}

func NewPaymentPlans(paymentPlans []models.PaymentPlan, count int64) interface{} {
	type paymentPlansResp struct {
		Base
		Count        int64         `json:"count"`
		PaymentPlans []paymentPlan `json:"paymentPlans"`
	}

	paymentPlanDTOs := make([]paymentPlan, len(paymentPlans))

	for i, plan := range paymentPlans {

		paymentPlanDTO := paymentPlan{
			Id:           plan.Id.String(),
			NextExecute:  plan.NextExecute,
			Name:         plan.Name,
			Schedule:     plan.Schedule,
			ScheduleUnit: plan.ScheduleUnit,
			Amount:       plan.Amount,
			GroupId:      plan.GroupId.String(),
		}

		if plan.ReceiverIsBank {
			paymentPlanDTO.ReceiverId = "bank"
		} else {
			paymentPlanDTO.ReceiverId = plan.ReceiverId.String()
		}

		if plan.SenderIsBank {
			paymentPlanDTO.SenderId = "bank"
		} else {
			paymentPlanDTO.SenderId = plan.SenderId.String()
		}

		paymentPlanDTOs[i] = paymentPlanDTO
	}

	return paymentPlansResp{
		Base: Base{
			Success: true,
		},
		Count:        count,
		PaymentPlans: paymentPlanDTOs,
	}
}
