package main

import (
	"log"

	"github.com/adrg/xdg"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/db"
	"github.com/Bananenpro/hbank-api/models"
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

func main() {
	config.Load([]string{"config.json", xdg.ConfigHome + "/hbank/config.json"})

	database, err := db.NewSqlite("database.sqlite")
	if err != nil {
		log.Fatalln("Couldn't connect to database:", err)
	}
	err = db.AutoMigrate(database)
	if err != nil {
		log.Fatalln("Couldn't auto migrate database:", err)
	}

	us := db.NewUserStore(database)
	gs := db.NewGroupStore(database)

	ExecutePaymentPlans(us, gs)

	log.Println("Done.")
}
