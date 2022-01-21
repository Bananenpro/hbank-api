package main

import (
	"time"

	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/services"
)

func ExecutePaymentPlan(userStore models.UserStore, groupStore models.GroupStore, paymentPlan *models.PaymentPlan) error {
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
