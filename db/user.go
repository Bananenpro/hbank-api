package db

import (
	"errors"
	"time"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/services"
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

func (us *UserStore) GetAll(except *models.User, page, pageSize int, descending bool) ([]models.User, error) {
	var users []models.User
	var err error

	order := "ASC"
	if descending {
		order = "DESC"
	}

	if except == nil {
		if page < 0 || pageSize < 0 {
			err = us.db.Omit("profile_picture").Order("name " + order).Find(&users).Error
		} else {
			err = us.db.Omit("profile_picture").Order("name " + order).Offset(page * pageSize).Limit(pageSize).Find(&users).Error
		}
	} else {
		if page < 0 || pageSize < 0 {
			err = us.db.Omit("profile_picture").Not("id = ?", except.Id).Order("name " + order).Find(&users).Error
		} else {
			err = us.db.Omit("profile_picture").Not("id = ?", except.Id).Order("name " + order).Offset(page * pageSize).Limit(pageSize).Find(&users).Error
		}
	}
	return users, err
}

func (us *UserStore) GetById(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := us.db.Omit("profile_picture").First(&user, id).Error
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
	err := us.db.Omit("profile_picture").First(&user, "email = ?", email).Error
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
	us.db.Delete(&models.ConfirmEmailLastSent{}, "email = ?", user.Email)
	us.db.Delete(&models.ForgotPasswordEmailLastSent{}, "email = ?", user.Email)
	us.db.Delete(&models.RefreshToken{}, "user_id = ?", user.Id)
	us.db.Delete(&models.PasswordToken{}, "user_id = ?", user.Id)
	us.db.Delete(&models.TwoFAToken{}, "user_id = ?", user.Id)
	us.db.Delete(&models.RecoveryCode{}, "user_id = ?", user.Id)
	us.db.Delete(&models.CashLogEntry{}, "user_id = ?", user.Id)
	us.db.Delete(&models.GroupInvitation{}, "user_id = ?", user.Id)
	us.db.Delete(&models.GroupMembership{}, "user_id = ?", user.Id)
	return us.db.Delete(user).Error
}

func (us *UserStore) DeleteById(id uuid.UUID) error {
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

func (us *UserStore) GetProfilePicture(user *models.User) ([]byte, error) {
	var u models.User
	err := us.db.Select("profile_picture").First(&u, user.Id).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}
	return u.ProfilePicture, nil
}

func (us *UserStore) GetConfirmEmailCode(user *models.User) (*models.ConfirmEmailCode, error) {
	var emailCode models.ConfirmEmailCode
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

func (us *UserStore) DeleteConfirmEmailCode(emailCode *models.ConfirmEmailCode) error {
	return us.db.Delete(emailCode).Error
}

func (us *UserStore) GetResetPasswordCode(user *models.User) (*models.ResetPasswordCode, error) {
	var emailCode models.ResetPasswordCode
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

func (us *UserStore) DeleteResetPasswordCode(emailCode *models.ResetPasswordCode) error {
	return us.db.Delete(emailCode).Error
}

func (us *UserStore) GetChangeEmailCode(user *models.User) (*models.ChangeEmailCode, error) {
	var emailCode models.ChangeEmailCode
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

func (us *UserStore) DeleteChangeEmailCode(emailCode *models.ChangeEmailCode) error {
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
		CodeHash:       hash,
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
	err := us.db.First(&token, "user_id = ? AND code_hash = ?", user.Id, services.HashToken(code)).Error

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
	err := us.db.First(&token, "user_id = ? AND code_hash = ?", user.Id, services.HashToken(code)).Error

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
	err := us.db.First(&rCode, "user_id = ? AND code_hash = ?", user.Id, services.HashToken(code)).Error

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

func (us *UserStore) NewRecoveryCodes(user *models.User) ([]string, error) {
	err := us.db.Where("user_id = ?", user.Id).Delete(&models.RecoveryCode{}).Error
	if err != nil {
		return []string{}, err
	}

	codes := make([]models.RecoveryCode, 10)
	codesStr := make([]string, 10)

	for i := range codes {
		codesStr[i] = services.GenerateRandomString(32)
		codes[i].CodeHash = services.HashToken(codesStr[i])
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

func (us *UserStore) GetCashLog(user *models.User, page, pageSize int, oldestFirst bool) ([]models.CashLogEntry, error) {
	var cashLog []models.CashLogEntry
	var err error
	if page < 0 || pageSize < 0 {
		if oldestFirst {
			err = us.db.Where("user_id = ?", user.Id).Order("created ASC").Find(&cashLog).Error
		} else {
			err = us.db.Where("user_id = ?", user.Id).Order("created DESC").Find(&cashLog).Error
		}
	} else {
		offset := page * pageSize
		if oldestFirst {
			err = us.db.Where("user_id = ?", user.Id).Order("created ASC").Offset(offset).Limit(pageSize).Find(&cashLog).Error
		} else {
			err = us.db.Where("user_id = ?", user.Id).Order("created DESC").Offset(offset).Limit(pageSize).Find(&cashLog).Error
		}
	}

	return cashLog, err
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

func (us *UserStore) GetCashLogEntryById(user *models.User, id uuid.UUID) (*models.CashLogEntry, error) {
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
