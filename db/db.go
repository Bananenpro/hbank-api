package db

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
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

// returns db and the id of db file
func NewTestDB() (*gorm.DB, string, error) {
	id := uuid.NewString()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("%s.sqlite", id)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	return db, id, err
}

func DeleteTestDB(id string) {
	err := os.Remove(fmt.Sprintf("%s.sqlite", id))
	if err != nil {
		log.Fatalln("Failed to delete test database:", err)
	}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.CashLogEntry{},

		&models.Group{},
		&models.GroupMembership{},
		&models.GroupPicture{},
		&models.GroupInvitation{},
		&models.TransactionLogEntry{},
		&models.PaymentPlan{},
	)
}
