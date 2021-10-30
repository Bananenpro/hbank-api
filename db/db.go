package db

import (
	"log"
	"os"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSqlite(filepath string) (*gorm.DB, error) {
	if config.Data.DBVerbose {
		return gorm.Open(sqlite.Open(filepath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		return gorm.Open(sqlite.Open(filepath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}
}

func NewTestDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open("database_test.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func DeleteTestDB() {
	err := os.Remove("database_test.sqlite")
	if err != nil {
		log.Fatalln("Failed to delete test database:", err)
	}
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
		&models.GroupMembership{},
		&models.GroupInvitation{},
		&models.TransactionLogEntry{},
		&models.PaymentPlan{},
	)
}
