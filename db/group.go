package db

import (
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
	return gs.db.Updates(group).Error
}

func (gs *GroupStore) Delete(group *models.Group) error {
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

func (gs *GroupStore) GetMembers(group *models.Group, page int, pageSize int, descending bool) ([]models.User, error) {
	var memberships []models.GroupMembership
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(group).Order("user_name "+order).Association("Memberships").Find(&memberships, "is_member = ?", true)
	} else {
		err = gs.db.Model(group).Order("user_name "+order).Offset(page*pageSize).Limit(pageSize).Association("Memberships").Find(&memberships, "is_member = ?", true)
	}

	userIds := make([]uuid.UUID, len(memberships))
	for i, m := range memberships {
		userIds[i] = m.UserId
	}

	var members []models.User
	err = gs.db.Omit("profile_picture").Order("name "+order).Find(&members, "id IN ?", userIds).Error

	return members, err
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

	if membership.IsAdmin {
		membership.IsMember = false
		err = gs.db.Select("is_member").Updates(&membership).Error
	} else {
		err = gs.db.Delete(&membership).Error
	}

	return err
}

func (gs *GroupStore) GetAdmins(group *models.Group, page int, pageSize int, descending bool) ([]models.User, error) {
	var memberships []models.GroupMembership
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(group).Order("user_name "+order).Association("Memberships").Find(&memberships, "is_admin = ?", true)
	} else {
		err = gs.db.Model(group).Order("user_name "+order).Offset(page*pageSize).Limit(pageSize).Association("Memberships").Find(&memberships, "is_admin = ?", true)
	}

	userIds := make([]uuid.UUID, len(memberships))
	for i, m := range memberships {
		userIds[i] = m.UserId
	}

	var members []models.User
	err = gs.db.Omit("profile_picture").Order("name "+order).Find(&members, "id IN ?", userIds).Error

	return members, err
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
