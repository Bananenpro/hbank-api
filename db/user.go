package db

import (
	"errors"
	"time"

	"github.com/Bananenpro/hbank2-api/config"
	"github.com/Bananenpro/hbank2-api/models"
	"github.com/Bananenpro/hbank2-api/services"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func (us *UserStore) Delete(user *models.User) error {
	return us.db.Delete(user).Error
}

func (us *UserStore) DeleteById(id uuid.UUID) error {
	return us.db.Delete(models.User{}, id).Error
}

func (us *UserStore) DeleteByEmail(email string) error {
	return us.db.Delete(models.User{}, "email = ?", email).Error
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

func (us *UserStore) GetRefreshToken(user *models.User, id uuid.UUID) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := us.db.First(&token, "user_id = ? AND id = ?", user.Id, id).Error
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

func (us *UserStore) GetRefreshTokens(user *models.User) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := us.db.Find(&tokens, "user_id = ?", user.Id).Error
	return tokens, err
}

func (us *UserStore) AddRefreshToken(user *models.User, refreshToken *models.RefreshToken) error {
	return us.db.Model(&user).Association("RefreshTokens").Append(refreshToken)
}

func (us *UserStore) RotateRefreshToken(user *models.User, oldRefreshToken *models.RefreshToken) (*models.RefreshToken, string, error) {
	if oldRefreshToken.UserId.String() != user.Id.String() {
		return nil, "", errors.New("Refresh-token doesn't belong to user")
	}
	oldRefreshToken.Used = true
	err := us.db.Model(oldRefreshToken).Select("used").Updates(oldRefreshToken).Error
	if err != nil {
		return nil, "", err
	}

	code := services.GenerateRandomString(64)
	hash, err := bcrypt.GenerateFromPassword([]byte(code), config.Data.BcryptCost)
	if err != nil {
		return nil, "", err
	}
	newRefreshToken := &models.RefreshToken{
		Code:           hash,
		ExpirationTime: time.Now().Unix() + config.Data.RefreshTokenLifetime,
		UserId:         user.Id,
	}

	err = us.db.Create(newRefreshToken).Error

	return newRefreshToken, code, err
}

func (us *UserStore) DeleteRefreshToken(refreshToken *models.RefreshToken) error {
	return us.db.Delete(&refreshToken).Error
}

func (us *UserStore) DeleteRefreshTokens(user *models.User) error {
	return us.db.Delete(models.RefreshToken{}, "user_id = ?", user.Id).Error
}

func (us *UserStore) GetPasswordTokenByCode(user *models.User, code string) (*models.PasswordToken, error) {
	var token models.PasswordToken
	err := us.db.First(&token, "user_id = ? AND code = ?", user.Id, services.HashToken(code)).Error

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

func (us *UserStore) GetPasswordTokens(user *models.User) ([]models.PasswordToken, error) {
	var tokens []models.PasswordToken
	err := us.db.Find(&tokens, "user_id = ?", user.Id).Error
	return tokens, err
}

func (us *UserStore) DeletePasswordToken(token *models.PasswordToken) error {
	return us.db.Delete(&token).Error
}

func (us *UserStore) GetTwoFATokenByCode(user *models.User, code string) (*models.TwoFAToken, error) {
	var token models.TwoFAToken
	err := us.db.First(&token, "user_id = ? AND code = ?", user.Id, services.HashToken(code)).Error

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

func (us *UserStore) GetTwoFATokens(user *models.User) ([]models.TwoFAToken, error) {
	var tokens []models.TwoFAToken
	err := us.db.Find(&tokens, "user_id = ?", user.Id).Error
	return tokens, err
}

func (us *UserStore) DeleteTwoFAToken(token *models.TwoFAToken) error {
	return us.db.Delete(&token).Error
}

func (us *UserStore) GetRecoveryCodeByCode(user *models.User, code string) (*models.RecoveryCode, error) {
	var rCode models.RecoveryCode
	err := us.db.First(&rCode, "user_id = ? AND code = ?", user.Id, services.HashToken(code)).Error

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

func (us *UserStore) NewRecoveryCodes(user *models.User) ([]string, error) {
	err := us.db.Where("user_id = ?", user.Id).Delete(&models.RecoveryCode{}).Error
	if err != nil {
		return []string{}, err
	}

	codes := make([]models.RecoveryCode, 10)
	codesStr := make([]string, 10)

	for i := range codes {
		codesStr[i] = services.GenerateRandomString(32)
		codes[i].Code = services.HashToken(codesStr[i])
	}

	user.RecoveryCodes = codes
	err = us.Update(user)

	return codesStr, err
}

func (us *UserStore) DeleteRecoveryCode(code *models.RecoveryCode) error {
	return us.db.Delete(code).Error
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

func (us *UserStore) GetForgotPasswordEmailLastSent(email string) (int64, error) {
	var lastSent models.ForgotPasswordEmailLastSent
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

func (us *UserStore) SetForgotPasswordEmailLastSent(email string, time int64) error {
	var lastSent models.ForgotPasswordEmailLastSent
	err := us.db.First(&lastSent, "email = ?", email).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			us.db.Create(&models.ForgotPasswordEmailLastSent{
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
