package db

import (
	"time"

	"github.com/Bananenpro/hbank-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GroupStore struct {
	db *gorm.DB
}

func NewGroupStore(db *gorm.DB) *GroupStore {
	return &GroupStore{
		db: db,
	}
}

func (gs *GroupStore) GetAllByUser(user *models.User, page int, pageSize int, descending bool) ([]models.Group, error) {
	var memberships []models.GroupMembership
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(user).Order("group_name " + order).Association("GroupMemberships").Find(&memberships)
	} else {
		err = gs.db.Model(user).Order("group_name " + order).Offset(page * pageSize).Limit(pageSize).Association("GroupMemberships").Find(&memberships)
	}

	if err != nil {
		return nil, err
	}

	groupIds := make([]uuid.UUID, len(memberships))
	for i, m := range memberships {
		groupIds[i] = m.GroupId
	}

	var groups []models.Group
	err = gs.db.Omit("group_picture").Order("name "+order).Find(&groups, "id IN ?", groupIds).Error

	return groups, err
}

func (gs *GroupStore) Count(user *models.User) (int64, error) {
	var count int64
	err := gs.db.Model(&models.GroupMembership{}).Where("user_id = ?", user.Id).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetById(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	err := gs.db.Omit("group_picture").First(&group, id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}
	return &group, nil
}

func (gs *GroupStore) Create(group *models.Group) error {
	return gs.db.Create(group).Error
}

func (gs *GroupStore) Update(group *models.Group) error {
	if group.Description == "" {
		err := gs.db.Select("description").Updates(group).Error
		if err != nil {
			return err
		}
	}
	return gs.db.Updates(group).Error
}

func (gs *GroupStore) UpdateGroupPicture(group *models.Group) error {
	return gs.db.Select("group_picture").Select("group_picture_id").Updates(group).Error
}

func (gs *GroupStore) Delete(group *models.Group) error {
	gs.db.Delete(&models.GroupInvitation{}, "group_id = ?", group.Id)
	gs.db.Delete(&models.GroupMembership{}, "group_id = ?", group.Id)
	gs.db.Delete(&models.TransactionLogEntry{}, "group_id = ?", group.Id)
	gs.db.Delete(&models.PaymentPlan{}, "group_id = ?", group.Id)
	return gs.db.Delete(group).Error
}

func (gs *GroupStore) DeleteById(id uuid.UUID) error {
	group, err := gs.GetById(id)
	if err != nil {
		return err
	}

	if group != nil {
		return gs.Delete(group)
	}

	return nil
}

func (gs *GroupStore) GetGroupPicture(group *models.Group) ([]byte, error) {
	var g models.Group
	err := gs.db.Select("group_picture").First(&g, group.Id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}
	return g.GroupPicture, nil
}

func (gs *GroupStore) GetMembers(except *models.User, searchInput string, group *models.Group, page int, pageSize int, descending bool) ([]models.User, error) {
	var memberships []models.GroupMembership
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if except == nil {
		except = &models.User{}
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(group).Order("user_name "+order).Not("user_id = ?", except.Id).Association("Memberships").Find(&memberships, "is_member = ? AND user_name LIKE ?", true, "%"+searchInput+"%")
	} else {
		err = gs.db.Model(group).Order("user_name "+order).Not("user_id = ?", except.Id).Offset(page*pageSize).Limit(pageSize).Association("Memberships").Find(&memberships, "is_member = ?  AND user_name LIKE ?", true, "%"+searchInput+"%")
	}

	userIds := make([]uuid.UUID, len(memberships))
	for i, m := range memberships {
		userIds[i] = m.UserId
	}

	var members []models.User
	err = gs.db.Omit("profile_picture").Order("name "+order).Find(&members, "id IN ?", userIds).Error

	return members, err
}

func (gs *GroupStore) MemberCount(group *models.Group) (int64, error) {
	var count int64
	err := gs.db.Model(&models.GroupMembership{}).Where("group_id = ? AND is_member = ?", group.Id, true).Count(&count).Error
	return count, err
}

func (gs *GroupStore) IsMember(group *models.Group, user *models.User) (bool, error) {
	err := gs.db.First(&models.GroupMembership{}, "group_id = ? AND user_id = ? AND is_member = ?", group.Id, user.Id, true).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (gs *GroupStore) AddMember(group *models.Group, user *models.User) error {
	var membership models.GroupMembership
	err := gs.db.First(&membership, "group_id = ? AND user_id = ?", group.Id, user.Id).Error
	if err == gorm.ErrRecordNotFound {
		err = gs.db.Model(group).Select("is_member").Association("Memberships").Append(&models.GroupMembership{
			IsMember:  true,
			GroupId:   group.Id,
			GroupName: group.Name,
			UserId:    user.Id,
			UserName:  user.Name,
		})
	} else if err == nil {
		membership.IsMember = true
		err = gs.db.Updates(&membership).Error
	}

	return err
}

func (gs *GroupStore) RemoveMember(group *models.Group, user *models.User) error {
	var membership models.GroupMembership
	err := gs.db.First(&membership, "group_id = ? AND user_id = ?", group.Id, user.Id).Error
	if err != nil {
		return err
	}

	gs.db.Where("group_id = ? AND sender_id = ?", group.Id, user.Id).Or("group_id = ? AND receiver_id = ?", group.Id, user.Id).Delete(&models.PaymentPlan{})

	if membership.IsAdmin {
		membership.IsMember = false
		err = gs.db.Select("is_member").Updates(&membership).Error
	} else {
		err = gs.db.Delete(&membership).Error
	}

	return err
}

func (gs *GroupStore) GetAdmins(except *models.User, searchInput string, group *models.Group, page int, pageSize int, descending bool) ([]models.User, error) {
	var memberships []models.GroupMembership
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if except == nil {
		except = &models.User{}
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(group).Order("user_name "+order).Not("user_id = ?", except.Id).Association("Memberships").Find(&memberships, "is_admin = ? AND user_name LIKE ?", true, "%"+searchInput+"%")
	} else {
		err = gs.db.Model(group).Order("user_name "+order).Not("user_id = ?", except.Id).Offset(page*pageSize).Limit(pageSize).Association("Memberships").Find(&memberships, "is_admin = ? AND user_name LIKE ?", true, "%"+searchInput+"%")
	}

	userIds := make([]uuid.UUID, len(memberships))
	for i, m := range memberships {
		userIds[i] = m.UserId
	}

	var members []models.User
	err = gs.db.Omit("profile_picture").Order("name "+order).Find(&members, "id IN ?", userIds).Error

	return members, err
}

func (gs *GroupStore) AdminCount(group *models.Group) (int64, error) {
	var count int64
	err := gs.db.Model(&models.GroupMembership{}).Where("group_id = ? AND is_admin = ?", group.Id, true).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetMemberships(except *models.User, searchInput string, group *models.Group, page int, pageSize int, descending bool) ([]models.GroupMembership, error) {
	var memberships []models.GroupMembership
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if except == nil {
		except = &models.User{}
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(group).Order("user_name "+order).Not("user_id = ?", except.Id).Association("Memberships").Find(&memberships, "user_name LIKE ?", "%"+searchInput+"%")
	} else {
		err = gs.db.Model(group).Order("user_name "+order).Not("user_id = ?", except.Id).Offset(page*pageSize).Limit(pageSize).Association("Memberships").Find(&memberships, "user_name LIKE ?", "%"+searchInput+"%")
	}

	return memberships, err
}

func (gs *GroupStore) MembershipCount(group *models.Group) (int64, error) {
	var count int64
	err := gs.db.Model(&models.GroupMembership{}).Where("group_id = ?", group.Id).Count(&count).Error
	return count, err
}

func (gs *GroupStore) IsAdmin(group *models.Group, user *models.User) (bool, error) {
	err := gs.db.First(&models.GroupMembership{}, "group_id = ? AND user_id = ? AND is_admin = ?", group.Id, user.Id, true).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (gs *GroupStore) AddAdmin(group *models.Group, user *models.User) error {
	var membership models.GroupMembership
	err := gs.db.First(&membership, "group_id = ? AND user_id = ?", group.Id, user.Id).Error
	if err == gorm.ErrRecordNotFound {
		err = gs.db.Model(group).Association("Memberships").Append(&models.GroupMembership{
			IsAdmin:   true,
			GroupId:   group.Id,
			GroupName: group.Name,
			UserId:    user.Id,
			UserName:  user.Name,
		})
	} else if err == nil {
		membership.IsAdmin = true
		err = gs.db.Select("is_admin").Updates(&membership).Error
	}

	return err
}

func (gs *GroupStore) RemoveAdmin(group *models.Group, user *models.User) error {
	var membership models.GroupMembership
	err := gs.db.First(&membership, "group_id = ? AND user_id = ?", group.Id, user.Id).Error
	if err != nil {
		return err
	}

	if membership.IsMember {
		membership.IsAdmin = false
		err = gs.db.Select("is_admin").Updates(&membership).Error
	} else {
		err = gs.db.Delete(&membership).Error
	}

	return err
}

func (gs *GroupStore) IsInGroup(group *models.Group, user *models.User) (bool, error) {
	err := gs.db.Where("group_id = ? AND user_id = ? AND is_member = ?", group.Id, user.Id, true).Or("group_id = ? AND user_id = ? AND is_admin = ?", group.Id, user.Id, true).First(&models.GroupMembership{}).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (gs *GroupStore) GetUserCount(group *models.Group) (int64, error) {
	count := int64(0)
	err := gs.db.Model(&models.GroupMembership{}).Where("group_id = ? AND is_member = ?", group.Id, true).Or("group_id = ? AND is_admin = ?", group.Id, true).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetTransactionLog(group *models.Group, user *models.User, searchInput string, page, pageSize int, oldestFirst bool) ([]models.TransactionLogEntry, error) {
	var log []models.TransactionLogEntry
	var err error

	order := "DESC"
	if oldestFirst {
		order = "ASC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Order("created "+order).Where("group_id = ? AND sender_id = ? AND title LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Or("group_id = ? AND receiver_id = ? AND title LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Find(&log).Error
	} else {
		err = gs.db.Order("created "+order).Offset(page*pageSize).Limit(pageSize).Where("group_id = ? AND sender_id = ? AND title LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Or("group_id = ? AND receiver_id = ? AND title LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Find(&log).Error
	}

	return log, err
}

func (gs *GroupStore) TransactionLogEntryCount(group *models.Group, user *models.User) (int64, error) {
	var count int64
	err := gs.db.Model(&models.TransactionLogEntry{}).Where("group_id = ? AND sender_id = ?", group.Id, user.Id).Or("group_id = ? AND receiver_id = ?", group.Id, user.Id).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetBankTransactionLog(group *models.Group, searchInput string, page, pageSize int, oldestFirst bool) ([]models.TransactionLogEntry, error) {
	var log []models.TransactionLogEntry
	var err error

	order := "DESC"
	if oldestFirst {
		order = "ASC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Order("created "+order).Where("group_id = ? AND sender_is_bank = ? AND title LIKE ?", group.Id, true, "%"+searchInput+"%").Or("group_id = ? AND receiver_is_bank = ? AND title LIKE ?", group.Id, true, "%"+searchInput+"%").Find(&log).Error
	} else {
		err = gs.db.Order("created "+order).Offset(page*pageSize).Limit(pageSize).Where("group_id = ? AND sender_is_bank = ? AND title LIKE ?", group.Id, true, "%"+searchInput+"%").Or("group_id = ? AND receiver_is_bank = ? AND title LIKE ?", group.Id, true, "%"+searchInput+"%").Find(&log).Error
	}

	return log, err
}

func (gs *GroupStore) BankTransactionLogEntryCount(group *models.Group) (int64, error) {
	var count int64
	err := gs.db.Model(&models.TransactionLogEntry{}).Where("group_id = ? AND sender_is_bank = ?", group.Id, true).Or("group_id = ? AND receiver_is_bank = ?", group.Id, true).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetTransactionLogEntryById(group *models.Group, id uuid.UUID) (*models.TransactionLogEntry, error) {
	var entry models.TransactionLogEntry
	err := gs.db.First(&entry, "group_id = ? AND id = ?", group.Id, id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &entry, nil
}

func (gs *GroupStore) GetLastTransactionLogEntry(group *models.Group, user *models.User) (*models.TransactionLogEntry, error) {
	var entry models.TransactionLogEntry
	err := gs.db.Order("created DESC").Where("group_id = ? AND sender_id = ?", group.Id, user.Id).Or("group_id = ? AND receiver_id = ?", group.Id, user.Id).First(&entry).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}
	return &entry, nil
}

func (gs *GroupStore) GetUserBalance(group *models.Group, user *models.User) (int, error) {
	lastLogEntry, err := gs.GetLastTransactionLogEntry(group, user)
	if err != nil {
		return 0, err
	}
	if lastLogEntry == nil {
		return 0, nil
	}

	if lastLogEntry.SenderId.String() == user.Id.String() {
		return lastLogEntry.NewBalanceSender, nil
	} else {
		return lastLogEntry.NewBalanceReceiver, nil
	}
}

func (gs *GroupStore) CreateTransaction(group *models.Group, senderIsBank, receiverIsBank bool, sender *models.User, receiver *models.User, title, description string, amount int) (*models.TransactionLogEntry, error) {
	return gs.CreateTransactionFromPaymentPlan(group, senderIsBank, receiverIsBank, sender, receiver, title, description, amount, uuid.UUID{})
}

func (gs *GroupStore) CreateTransactionFromPaymentPlan(group *models.Group, senderIsBank, receiverIsBank bool, sender *models.User, receiver *models.User, title, description string, amount int, paymentPlanId uuid.UUID) (*models.TransactionLogEntry, error) {
	var err error

	oldBalanceSender := 0
	newBalanceSender := 0
	if !senderIsBank {
		oldBalanceSender, err = gs.GetUserBalance(group, sender)
		if err != nil {
			return nil, err
		}
		newBalanceSender = oldBalanceSender - amount
	}

	oldBalanceReceiver := 0
	newBalanceReceiver := 0
	if !receiverIsBank {
		oldBalanceReceiver, err = gs.GetUserBalance(group, receiver)
		if err != nil {
			return nil, err
		}
		newBalanceReceiver = oldBalanceReceiver + amount
	}

	senderId := uuid.UUID{}
	if !senderIsBank {
		senderId = sender.Id
	}

	receiverId := uuid.UUID{}
	if !receiverIsBank {
		receiverId = receiver.Id
	}

	transaction := models.TransactionLogEntry{
		Title:       title,
		Description: description,
		Amount:      int(amount),
		GroupId:     group.Id,

		SenderIsBank:            senderIsBank,
		SenderId:                senderId,
		BalanceDifferenceSender: -amount,
		NewBalanceSender:        newBalanceSender,

		ReceiverIsBank:            receiverIsBank,
		ReceiverId:                receiverId,
		BalanceDifferenceReceiver: amount,
		NewBalanceReceiver:        newBalanceReceiver,

		PaymentPlanId: paymentPlanId,
	}

	err = gs.db.Create(&transaction).Error

	return &transaction, err
}

func (gs *GroupStore) CreateInvitation(group *models.Group, user *models.User, message string) (*models.GroupInvitation, error) {
	invitation := &models.GroupInvitation{
		Message:   message,
		GroupName: group.Name,
		GroupId:   group.Id,
		UserId:    user.Id,
	}

	err := gs.db.Create(invitation).Error

	return invitation, err
}

func (gs *GroupStore) GetInvitationById(id uuid.UUID) (*models.GroupInvitation, error) {
	var invitation models.GroupInvitation
	err := gs.db.First(&invitation, id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &invitation, nil
}

func (gs *GroupStore) GetInvitationsByGroup(group *models.Group, page, pageSize int, oldestFirst bool) ([]models.GroupInvitation, error) {
	order := "DESC"
	if oldestFirst {
		order = "ASC"
	}

	var invitations []models.GroupInvitation
	var err error
	if page < 0 || pageSize < 0 {
		err = gs.db.Order("created "+order).Find(&invitations, "group_id = ?", group.Id).Error
	} else {
		err = gs.db.Order("created "+order).Offset(page*pageSize).Limit(pageSize).Find(&invitations, "group_id = ?", group.Id).Error
	}

	return invitations, err
}

func (gs *GroupStore) InvitationCountByGroup(group *models.Group) (int64, error) {
	var count int64
	err := gs.db.Model(&models.GroupInvitation{}).Where("group_id = ?", group.Id).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetInvitationsByUser(user *models.User, page, pageSize int, oldestFirst bool) ([]models.GroupInvitation, error) {
	order := "DESC"
	if oldestFirst {
		order = "ASC"
	}

	var invitations []models.GroupInvitation
	var err error
	if page < 0 || pageSize < 0 {
		err = gs.db.Order("created "+order).Find(&invitations, "user_id = ?", user.Id).Error
	} else {
		err = gs.db.Order("created "+order).Offset(page*pageSize).Limit(pageSize).Find(&invitations, "user_id = ?", user.Id).Error
	}

	return invitations, err
}

func (gs *GroupStore) InvitationCountByUser(user *models.User) (int64, error) {
	var count int64
	err := gs.db.Model(&models.GroupInvitation{}).Where("user_id = ?", user.Id).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetInvitationByGroupAndUser(group *models.Group, user *models.User) (*models.GroupInvitation, error) {
	var invitation models.GroupInvitation
	err := gs.db.First(&invitation, "group_id = ? AND user_id = ?", group.Id, user.Id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &invitation, nil
}

func (gs *GroupStore) DeleteInvitation(invitation *models.GroupInvitation) error {
	return gs.db.Delete(invitation).Error
}

func (gs *GroupStore) GetPaymentPlans(group *models.Group, user *models.User, searchInput string, page, pageSize int, descending bool) ([]models.PaymentPlan, error) {
	var paymentPlans []models.PaymentPlan
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Order("next_execute "+order).Where("group_id = ? AND sender_id = ? AND name LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Or("group_id = ? AND receiver_id = ? AND name LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Find(&paymentPlans).Error
	} else {
		err = gs.db.Order("next_execute "+order).Offset(page*pageSize).Limit(pageSize).Where("group_id = ? AND sender_id = ? AND name LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Or("group_id = ? AND receiver_id = ? AND name LIKE ?", group.Id, user.Id, "%"+searchInput+"%").Find(&paymentPlans).Error
	}

	return paymentPlans, err
}

func (gs *GroupStore) PaymentPlanCount(group *models.Group, user *models.User) (int64, error) {
	var count int64
	err := gs.db.Model(&models.PaymentPlan{}).Where("group_id = ? AND sender_id = ?", group.Id, user.Id).Or("group_id = ? AND receiver_id = ?", group.Id, user.Id).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetBankPaymentPlans(group *models.Group, searchInput string, page, pageSize int, descending bool) ([]models.PaymentPlan, error) {
	var paymentPlans []models.PaymentPlan
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Order("next_execute "+order).Where("group_id = ? AND sender_is_bank = ? AND name LIKE ?", group.Id, true, "%"+searchInput+"%").Or("group_id = ? AND receiver_is_bank = ? AND name LIKE ?", group.Id, true, "%"+searchInput+"%").Find(&paymentPlans).Error
	} else {
		err = gs.db.Order("next_execute "+order).Where("group_id = ? AND sender_is_bank = ? AND name LIKE ?", group.Id, true, "%"+searchInput+"%").Or("group_id = ? AND receiver_is_bank = ? AND name LIKE ?", group.Id, true, "%"+searchInput+"%").Offset(page * pageSize).Limit(pageSize).Find(&paymentPlans).Error
	}

	return paymentPlans, err
}

func (gs *GroupStore) BankPaymentPlanCount(group *models.Group) (int64, error) {
	var count int64
	err := gs.db.Model(&models.PaymentPlan{}).Where("group_id = ? AND sender_is_bank = ?", group.Id, true).Or("group_id = ? AND receiver_is_bank = ?", group.Id, true).Count(&count).Error
	return count, err
}

func (gs *GroupStore) GetPaymentPlansThatNeedToBeExecuted() ([]models.PaymentPlan, error) {
	var paymentPlans []models.PaymentPlan
	err := gs.db.Find(&paymentPlans, "next_execute <= ?", time.Now().Unix()).Error
	return paymentPlans, err
}

func (gs *GroupStore) GetPaymentPlanById(group *models.Group, id uuid.UUID) (*models.PaymentPlan, error) {
	var paymentPlan models.PaymentPlan
	err := gs.db.First(&paymentPlan, "group_id = ? AND id = ?", group.Id, id).Error

	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &paymentPlan, nil
}

func (gs *GroupStore) CreatePaymentPlan(group *models.Group, senderIsBank, receiverIsBank bool, sender *models.User, receiver *models.User, name, description string, amount, paymentCount, schedule int, scheduleUnit string, firstPayment int64) (*models.PaymentPlan, error) {
	paymentPlan := models.PaymentPlan{
		Name:           name,
		Description:    description,
		Amount:         amount,
		PaymentCount:   paymentCount,
		NextExecute:    firstPayment,
		Schedule:       schedule,
		ScheduleUnit:   scheduleUnit,
		SenderIsBank:   senderIsBank,
		ReceiverIsBank: receiverIsBank,
		GroupId:        group.Id,
	}

	if !senderIsBank {
		paymentPlan.SenderId = sender.Id
	}

	if !receiverIsBank {
		paymentPlan.ReceiverId = receiver.Id
	}

	err := gs.db.Create(&paymentPlan).Error

	return &paymentPlan, err
}

func (gs *GroupStore) UpdatePaymentPlan(paymentPlan *models.PaymentPlan) error {
	return gs.db.Updates(paymentPlan).Error
}

func (gs *GroupStore) DeletePaymentPlan(paymentPlan *models.PaymentPlan) error {
	gs.db.Model(&models.TransactionLogEntry{}).Where("payment_plan_id = ?", paymentPlan.Id).Update("payment_plan_id", uuid.UUID{})
	return gs.db.Delete(paymentPlan).Error
}
