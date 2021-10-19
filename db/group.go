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

func (gs *GroupStore) GetAllByMember(member *models.User, page int, pageSize int, descending bool) ([]models.Group, error) {
	var groups []models.Group
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(member).Order("name " + order).Omit("group_picture").Association("MemberGroups").Find(&groups)
	} else {
		err = gs.db.Model(member).Order("name " + order).Omit("group_picture").Offset(page * pageSize).Limit(pageSize).Association("MemberGroups").Find(&groups)
	}

	return groups, err
}

func (gs *GroupStore) GetAllByAdmin(admin *models.User, page int, pageSize int, descending bool) ([]models.Group, error) {
	var groups []models.Group
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(admin).Order("name " + order).Omit("group_picture").Association("AdminGroups").Find(&groups)
	} else {
		err = gs.db.Model(admin).Order("name " + order).Omit("group_picture").Offset(page * pageSize).Limit(pageSize).Association("AdminGroups").Find(&groups)
	}

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

func (gs *GroupStore) Create(user *models.User, group *models.Group) error {
	return gs.db.Model(user).Association("AdminGroups").Append(group)
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
	var members []models.User
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(group).Order("name " + order).Omit("profile_picture").Association("Members").Find(&members)
	} else {
		err = gs.db.Model(group).Order("name " + order).Omit("profile_picture").Offset(page * pageSize).Limit(pageSize).Association("Members").Find(&members)
	}

	return members, err
}

func (gs *GroupStore) IsMember(group *models.Group, user *models.User) (bool, error) {
	var members []models.User
	err := gs.db.Model(group).Omit("profile_picture").Limit(1).Association("Members").Find(&members, "id = ?", user.Id)

	return len(members) == 1, err
}

func (gs *GroupStore) AddMember(group *models.Group, user *models.User) error {
	return gs.db.Model(group).Association("Members").Append(user)
}

func (gs *GroupStore) RemoveMember(group *models.Group, user *models.User) error {
	return gs.db.Model(group).Association("Members").Delete(user)
}

func (gs *GroupStore) GetAdmins(group *models.Group, page int, pageSize int, descending bool) ([]models.User, error) {
	var admins []models.User
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = gs.db.Model(group).Order("name " + order).Omit("profile_picture").Association("Admins").Find(&admins)
	} else {
		err = gs.db.Model(group).Order("name " + order).Omit("profile_picture").Offset(page * pageSize).Limit(pageSize).Association("Admins").Find(&admins)
	}

	return admins, err
}

func (gs *GroupStore) IsAdmin(group *models.Group, user *models.User) (bool, error) {
	var admins []models.User
	err := gs.db.Model(group).Omit("profile_picture").Limit(1).Association("Admins").Find(&admins, "id = ?", user.Id)

	return len(admins) == 1, err
}

func (gs *GroupStore) AddAdmin(group *models.Group, user *models.User) error {
	return gs.db.Model(group).Association("Admins").Append(user)
}

func (gs *GroupStore) RemoveAdmin(group *models.Group, user *models.User) error {
	return gs.db.Model(group).Association("Admins").Delete(user)
}

func (gs *GroupStore) IsInGroup(group *models.Group, user *models.User) (bool, error) {
	isMember, err := gs.IsMember(group, user)
	if err != nil {
		return false, err
	}
	if isMember {
		return true, nil
	}

	isAdmin, err := gs.IsAdmin(group, user)

	return isAdmin, err
}
