package db

import (
	"github.com/Bananenpro/hbank2-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSqlite(filepath string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(filepath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
		&models.EmailCode{},
		&models.RefreshToken{},
		&models.PasswordToken{},
		&models.TwoFAToken{},
		&models.RecoveryCode{},
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
