package bindings

type DeleteUser struct {
	Password   string `json:"password"`
	TwoFAToken string `json:"two_fa_token"`
}

type DeleteUserByConfirmEmailCode struct {
	ConfirmEmailCode string `json:"code"`
}
