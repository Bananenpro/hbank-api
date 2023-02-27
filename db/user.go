package db

import (
	"gorm.io/gorm"

	"github.com/juho05/hbank-api/models"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (us *UserStore) GetAll(exclude []string, searchInput string, page, pageSize int, descending bool) ([]models.User, error) {
	var users []models.User
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if page < 0 || pageSize < 0 {
		err = us.db.Not(map[string]interface{}{"id": exclude}).Order("name "+order).Find(&users, "name LIKE ? AND publicly_visible = ?", "%"+searchInput+"%", true).Error
	} else {
		err = us.db.Not(map[string]interface{}{"id": exclude}).Order("name "+order).Offset(page*pageSize).Limit(pageSize).Find(&users, "name LIKE ? AND publicly_visible = ?", "%"+searchInput+"%", true).Error
	}

	return users, err
}

func (us *UserStore) Count() (int64, error) {
	var count int64
	err := us.db.Model(&models.User{}).Where("publicly_visible = ?", true).Count(&count).Error
	return count, err
}

func (us *UserStore) GetById(id string) (*models.User, error) {
	var user models.User
	err := us.db.First(&user, "id = ?", id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (us *UserStore) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := us.db.First(&user, "email = ?", email).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (us *UserStore) Create(user *models.User) error {
	return us.db.Create(user).Error
}

func (us *UserStore) Update(user *models.User) error {
	oldUser, err := us.GetById(user.Id)
	if err != nil {
		return err
	}
	if oldUser.Name != user.Name {
		us.db.Model(models.GroupMembership{}).Where("user_id = ?", user.Id).Update("user_name", user.Name)
	}
	return us.db.Select("*").Updates(user).Error
}

func (us *UserStore) Delete(user *models.User) error {
	us.db.Delete(&models.CashLogEntry{}, "user_id = ?", user.Id)
	us.db.Delete(&models.GroupInvitation{}, "user_id = ?", user.Id)
	us.db.Delete(&models.GroupMembership{}, "user_id = ?", user.Id)
	us.db.Where("sender_id = ?", user.Id).Or("receiver_id = ?", user.Id).Delete(&models.PaymentPlan{})
	return us.db.Delete(user).Error
}

func (us *UserStore) DeleteById(id string) error {
	user, err := us.GetById(id)
	if err != nil {
		return err
	}

	if user != nil {
		return us.Delete(user)
	}

	return nil
}

func (us *UserStore) DeleteByEmail(email string) error {
	user, err := us.GetByEmail(email)
	if err != nil {
		return err
	}

	if user != nil {
		return us.Delete(user)
	}

	return nil
}

func (us *UserStore) GetCashLog(user *models.User, searchInput string, page, pageSize int, oldestFirst bool) ([]models.CashLogEntry, error) {
	var cashLog []models.CashLogEntry
	var err error
	if page < 0 || pageSize < 0 {
		if oldestFirst {
			err = us.db.Where("user_id = ? AND change_title LIKE ?", user.Id, "%"+searchInput+"%").Order("created ASC").Find(&cashLog).Error
		} else {
			err = us.db.Where("user_id = ? AND change_title LIKE ?", user.Id, "%"+searchInput+"%").Order("created DESC").Find(&cashLog).Error
		}
	} else {
		offset := page * pageSize
		if oldestFirst {
			err = us.db.Where("user_id = ? AND change_title LIKE ?", user.Id, "%"+searchInput+"%").Order("created ASC").Offset(offset).Limit(pageSize).Find(&cashLog).Error
		} else {
			err = us.db.Where("user_id = ? AND change_title LIKE ?", user.Id, "%"+searchInput+"%").Order("created DESC").Offset(offset).Limit(pageSize).Find(&cashLog).Error
		}
	}

	return cashLog, err
}

func (us *UserStore) CashLogEntryCount(user *models.User) (int64, error) {
	var count int64
	err := us.db.Model(&models.CashLogEntry{}).Where("user_id = ?", user.Id).Count(&count).Error
	return count, err
}

func (us *UserStore) GetLastCashLogEntry(user *models.User) (*models.CashLogEntry, error) {
	var cashLog []models.CashLogEntry
	err := us.db.Where("user_id = ?", user.Id).Order("created desc").Limit(1).Find(&cashLog).Error
	if err != nil {
		return nil, err
	}

	if len(cashLog) == 0 {
		return nil, nil
	}

	return &cashLog[0], nil
}

func (us *UserStore) GetCashLogEntryById(user *models.User, id string) (*models.CashLogEntry, error) {
	var cashLogEntry models.CashLogEntry
	err := us.db.First(&cashLogEntry, "id = ? AND user_id = ?", id, user.Id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}
	return &cashLogEntry, nil
}

func (us *UserStore) AddCashLogEntry(user *models.User, entry *models.CashLogEntry) error {
	lastEntry, err := us.GetLastCashLogEntry(user)
	if err != nil {
		return err
	}

	totalAmount := 0

	totalAmount += 1 * entry.Ct1
	totalAmount += 2 * entry.Ct2
	totalAmount += 5 * entry.Ct5
	totalAmount += 10 * entry.Ct10
	totalAmount += 20 * entry.Ct20
	totalAmount += 50 * entry.Ct50

	totalAmount += 100 * entry.Eur1
	totalAmount += 200 * entry.Eur2
	totalAmount += 500 * entry.Eur5
	totalAmount += 1000 * entry.Eur10
	totalAmount += 2000 * entry.Eur20
	totalAmount += 5000 * entry.Eur50
	totalAmount += 10000 * entry.Eur100
	totalAmount += 20000 * entry.Eur200
	totalAmount += 50000 * entry.Eur500

	entry.TotalAmount = totalAmount

	if lastEntry != nil {
		entry.ChangeDifference = entry.TotalAmount - lastEntry.TotalAmount
	} else {
		entry.ChangeDifference = entry.TotalAmount
	}

	return us.db.Model(&user).Association("CashLog").Append(entry)
}
