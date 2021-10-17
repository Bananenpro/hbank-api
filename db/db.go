package db

import (
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSqlite(filepath string) (*gorm.DB, error) {
	if config.Data.Debug {
		return gorm.Open(sqlite.Open(filepath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		return gorm.Open(sqlite.Open(filepath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}
}

func NewInMemory() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.ConfirmEmailLastSent{},
		&models.ForgotPasswordEmailLastSent{},
		&models.ConfirmEmailCode{},
		&models.ResetPasswordCode{},
		&models.ChangeEmailCode{},
		&models.RefreshToken{},
		&models.PasswordToken{},
		&models.TwoFAToken{},
		&models.RecoveryCode{},
		&models.CashLogEntry{},

		&models.Group{},
	)
}

func Clear(db *gorm.DB) error {
	var users []models.User
	err := db.Find(&users).Error
	if err != nil {
		return err
	}

	err = db.Delete(&users).Error
	if err != nil {
		return err
	}

	return nil
}
