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
		&models.EmailCode{},
		&models.RefreshToken{},
		&models.LoginToken{},
		&models.RecoveryCode{},
	)
}
