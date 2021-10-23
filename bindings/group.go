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
}

type CreateInvitation struct {
	Message string
	UserId  string `json:"user_id" form:"user_id"`
}
