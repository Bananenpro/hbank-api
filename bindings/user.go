package bindings

type DeleteUser struct {
	Password   string `json:"password" form:"password"`
	TwoFAToken string `json:"twoFAToken" form:"twoFAToken"`
}

type DeleteUserByConfirmEmailCode struct {
	ConfirmEmailCode string `json:"code" form:"code"`
}

type UpdateUser struct {
	ProfilePicturePrivacy   string `json:"profilePicturePrivacy" form:"profilePicturePrivacy"`
	DontSendInvitationEmail bool   `json:"dontSendInvitationEmail" form:"dontSendInvitationEmail"`
}

type AddCashLogEntry struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	Ct1    uint `json:"ct1"`
	Ct2    uint `json:"ct2"`
	Ct5    uint `json:"ct5"`
	Ct10   uint `json:"ct10"`
	Ct20   uint `json:"ct20"`
	Ct50   uint `json:"ct50"`
	Eur1   uint `json:"eur1"`
	Eur2   uint `json:"eur2"`
	Eur5   uint `json:"eur5"`
	Eur10  uint `json:"eur10"`
	Eur20  uint `json:"eur20"`
	Eur50  uint `json:"eur50"`
	Eur100 uint `json:"eur100"`
	Eur200 uint `json:"eur200"`
	Eur500 uint `json:"eur500"`
}

type Id struct {
	Id string `json:"id"`
}
