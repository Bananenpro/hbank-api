package bindings

type CreateGroup struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	OnlyAdmin   bool   `json:"only_admin" form:"only_admin"`
}

type CreateTransaction struct {
	Title       string `json:"title" form:"title"`
	Description string `json:"description" form:"description"`
	Amount      uint   `json:"amount" form:"amount"`
	ReceiverId  string `json:"receiver_id" form:"receiver_id"`
	FromBank    bool   `json:"from_bank" form:"from_bank"`
}

type CreatePaymentPlan struct {
	Name         string `json:"name" form:"name"`
	Description  string `json:"description" form:"description"`
	Amount       uint   `json:"amount" form:"amount"`
	ReceiverId   string `json:"receiver_id" form:"receiver_id"`
	FromBank     bool   `json:"from_bank" form:"from_bank"`
	Schedule     uint   `json:"schedule" form:"schedule"`
	ScheduleUnit string `json:"schedule_unit" form:"schedule_unit"`
	// UTC date of first payment with format "YYYY-MM-DD"
	FirstPayment string `json:"first_payment"`
	// negative payment count for unlimited payments
	PaymentCount int `json:"payment_count"`
}

type UpdatePaymentPlan struct {
	Name         string `json:"name" form:"name"`
	Description  string `json:"description" form:"description"`
	Amount       uint   `json:"amount" form:"amount"`
	Schedule     uint   `json:"schedule" form:"schedule"`
	ScheduleUnit string `json:"schedule_unit" form:"schedule_unit"`
}

type CreateInvitation struct {
	Message string `json:"message" form:"message"`
	UserId  string `json:"user_id" form:"user_id"`
}
