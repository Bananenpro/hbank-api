package main

import (
	"log"
	"time"

	"github.com/juho05/h-bank/models"
	"github.com/juho05/h-bank/services"
)

var StopPaymentPlanTicker = make(chan struct{})

func StartPaymentPlanTicker(us models.UserStore, gs models.GroupStore) {
	log.Println("[payment-plans] Starting ticker...")
	ticker := time.NewTicker(time.Hour)
	go func() {
		for {
			executePaymentPlans(us, gs)
			select {
			case <-ticker.C:
				continue
			case <-StopPaymentPlanTicker:
				log.Println("[payment-plans] Stopping ticker...")
				ticker.Stop()
				return
			}
		}
	}()
}

func executePaymentPlans(us models.UserStore, gs models.GroupStore) {
	paymentPlans, err := gs.GetPaymentPlansThatNeedToBeExecuted()
	if err != nil {
		log.Println("[payment-plans] ERROR: Couldn't retrieve payment plans:", err)
		return
	}

	log.Printf("[payment-plans] Executing %d payment plans...", len(paymentPlans))

	for _, p := range paymentPlans {
		err = executePaymentPlan(us, gs, &p)
		if err != nil {
			log.Printf("[payment-plans] ERROR: Couldn't execute payment plan with id '%s': %s", p.Id, err)
		}
	}

	log.Println("[payment-plans] Done.")
}

func executePaymentPlan(userStore models.UserStore, groupStore models.GroupStore, paymentPlan *models.PaymentPlan) error {
	for paymentPlan.NextExecute <= time.Now().Unix() {
		group, err := groupStore.GetById(paymentPlan.GroupId)
		if err != nil {
			return err
		}
		if group == nil {
			return groupStore.Delete(group)
		}

		sender, err := userStore.GetById(paymentPlan.SenderId)
		if err != nil {
			return err
		}

		receiver, err := userStore.GetById(paymentPlan.ReceiverId)
		if err != nil {
			return err
		}

		if !paymentPlan.SenderIsBank {
			balance, err := groupStore.GetUserBalance(group, sender)
			if err != nil {
				return err
			}
			if balance-paymentPlan.Amount < 0 {
				break
			}
		}

		_, err = groupStore.CreateTransactionFromPaymentPlan(group, paymentPlan.SenderIsBank, paymentPlan.ReceiverIsBank, sender, receiver, paymentPlan.Name, paymentPlan.Description, paymentPlan.Amount, paymentPlan.Id)
		if err != nil {
			return err
		}

		paymentPlan.NextExecute = services.AddTime(paymentPlan.NextExecute, paymentPlan.Schedule, paymentPlan.ScheduleUnit)

		if paymentPlan.PaymentCount >= 0 {
			paymentPlan.PaymentCount -= 1

			if paymentPlan.PaymentCount <= 0 {
				return groupStore.DeletePaymentPlan(paymentPlan)
			}
		}

		err = groupStore.UpdatePaymentPlan(paymentPlan)
		if err != nil {
			return err
		}
	}

	return nil
}
