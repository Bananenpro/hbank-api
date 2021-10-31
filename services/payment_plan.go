package services

import (
	"log"
	"time"

	"github.com/Bananenpro/hbank-api/models"
)

func ExecutePaymentPlan(userStore models.UserStore, groupStore models.GroupStore, paymentPlan *models.PaymentPlan) error {
	if paymentPlan.NextExecute <= time.Now().Unix() {
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

		err = groupStore.CreateTransactionFromPaymentPlan(group, paymentPlan.SenderIsBank, paymentPlan.ReceiverIsBank, sender, receiver, paymentPlan.Name, paymentPlan.Description, paymentPlan.Amount, paymentPlan.Id)
		if err != nil {
			return err
		}

		paymentPlan.NextExecute = AddTime(paymentPlan.NextExecute, paymentPlan.Schedule, paymentPlan.ScheduleUnit)

		if paymentPlan.PaymentCount >= 0 {
			paymentPlan.PaymentCount -= 1

			if paymentPlan.PaymentCount <= 0 {
				return groupStore.DeletePaymentPlan(paymentPlan)
			}
		}

		return groupStore.UpdatePaymentPlan(paymentPlan)
	}

	return nil
}

func AddTime(unixTime int64, value int, unit string) int64 {
	t := time.Unix(unixTime, 0).UTC()
	switch unit {
	case models.ScheduleUnitDay:
		return t.AddDate(0, 0, value).Unix()
	case models.ScheduleUnitWeek:
		return t.AddDate(0, 0, value*7).Unix()
	case models.ScheduleUnitMonth:
		return t.AddDate(0, value, 0).Unix()
	case models.ScheduleUnitYear:
		return t.AddDate(value, 0, 0).Unix()
	default:
		log.Println("Error: unknown time unit:", unit)
		return 0
	}
}
