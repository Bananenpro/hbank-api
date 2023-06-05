package main

import (
	"fmt"
	"log"

	"github.com/adrg/xdg"

	"github.com/juho05/hbank-api/config"
	"github.com/juho05/hbank-api/db"
	"github.com/juho05/hbank-api/models"
)

func ExecutePaymentPlans(us models.UserStore, gs models.GroupStore) {
	paymentPlans, err := gs.GetPaymentPlansThatNeedToBeExecuted()
	if err != nil {
		log.Fatalln("Couldn't retrieve payment plans:", err)
	}

	log.Printf("Executing %d payment plans...", len(paymentPlans))

	for _, p := range paymentPlans {
		err = ExecutePaymentPlan(us, gs, &p)
		if err != nil {
			log.Printf("Couldn't execute payment plan with id '%s': %s", p.Id, err)
		}
	}
}

func run() error {
	config.Load([]string{"config.json", xdg.ConfigHome + "/hbank/config.json"})

	database, err := db.NewSqlite("database.sqlite?_pragma=foreign_keys(1)&_pragma=busy_timeout(3000)&_pragma=journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("Couldn't connect to database: %w", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("Failed to get generic DB interface: %w", err)
	}
	defer sqlDB.Close()
	err = db.AutoMigrate(database)
	if err != nil {
		return fmt.Errorf("Couldn't auto migrate database: %w", err)
	}

	us := db.NewUserStore(database)
	gs := db.NewGroupStore(database)

	ExecutePaymentPlans(us, gs)

	log.Println("Done.")
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
