package db

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gitlab.com/Bananenpro05/hbank2-api/config"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"gitlab.com/Bananenpro05/hbank2-api/services"
	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (us *UserStore) GetById(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := us.db.First(&user, id).Error
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
	return us.db.Updates(user).Error
}

func (us *UserStore) GetEmailCode(user *models.User) (*models.EmailCode, error) {
	var emailCode models.EmailCode
	err := us.db.First(&emailCode, "user_id = ?", user.Id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &emailCode, nil
}

func (us *UserStore) DeleteEmailCode(emailCode *models.EmailCode) error {
	return us.db.Delete(emailCode).Error
}

func (us *UserStore) GetRefreshToken(user *models.User, code string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := us.db.First(&token, "user_id = ? AND code = ?", user.Id, code).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &token, nil
}

func (us *UserStore) RotateRefreshToken(user *models.User, oldRefreshToken *models.RefreshToken) (*models.RefreshToken, error) {
	if oldRefreshToken.UserId != user.Id {
		return nil, errors.New("Refresh-token doesn't belong to user")
	}
	err := us.db.Delete(oldRefreshToken).Error
	if err != nil {
		return nil, err
	}

	newRefreshToken := &models.RefreshToken{
		Code:           services.GenerateRandomString(64),
		ExpirationTime: time.Now().UnixMilli() + config.Data.RefreshTokenLifetime,
		UserId:         user.Id,
	}

	err = us.db.Create(newRefreshToken).Error

	return newRefreshToken, err
}

func (us *UserStore) GetLoginTokenByCode(user *models.User, code string) (*models.LoginToken, error) {
	var token models.LoginToken
	err := us.db.First(&token, "user_id = ? AND code = ?", user.Id, code).Error

	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &token, nil
}

func (us *UserStore) GetLoginTokens(user *models.User) ([]models.LoginToken, error) {
	var tokens []models.LoginToken
	err := us.db.Find(&tokens, "user_id = ?", user.Id).Error
	return tokens, err
}

func (us *UserStore) DeleteLoginToken(token *models.LoginToken) error {
	return us.db.Delete(&token).Error
}

func (us *UserStore) GetRecoveryCodeByCode(user *models.User, code string) (*models.RecoveryCode, error) {
	var rCode models.RecoveryCode
	err := us.db.First(&rCode, "user_id = ? AND code = ?", user.Id, code).Error

	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &rCode, nil
}

func (us *UserStore) GetRecoveryCodes(user *models.User) ([]models.RecoveryCode, error) {
	var codes []models.RecoveryCode
	err := us.db.Find(&codes, "user_id = ?", user.Id).Error
	return codes, err
}

func (us *UserStore) NewRecoveryCodes(user *models.User) ([]models.RecoveryCode, error) {
	err := us.db.Where("user_id = ?", user.Id).Delete(&models.RecoveryCode{}).Error
	if err != nil {
		return []models.RecoveryCode{}, err
	}

	codes := make([]models.RecoveryCode, 10)

	for i := range codes {
		codes[i].Code = services.GenerateRandomString(64)
	}

	user.RecoveryCodes = codes
	err = us.Update(user)

	return codes, err
}

func (us *UserStore) GetConfirmEmailLastSent(email string) (int64, error) {
	var lastSent models.ConfirmEmailLastSent
	err := us.db.First(&lastSent, "email = ?", email).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return 0, nil
		default:
			return 0, err
		}
	}

	return lastSent.LastSent, nil
}

func (us *UserStore) SetConfirmEmailLastSent(email string, time int64) error {
	var lastSent models.ConfirmEmailLastSent
	err := us.db.First(&lastSent, "email = ?", email).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			us.db.Create(&models.ConfirmEmailLastSent{
				Email:    email,
				LastSent: time,
			})
			return nil
		default:
			return err
		}
	}

	lastSent.LastSent = time
	return us.db.Updates(&lastSent).Error
}
