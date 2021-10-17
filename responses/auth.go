package responses

type RegisterSuccess struct {
	Base
	UserId    string   `json:"user_id"`
	UserEmail string   `json:"user_email"`
	Codes     []string `json:"recovery_codes"`
}

type Token struct {
	Base
	Token string `json:"token"`
}

type RecoveryCodes struct {
	Base
	Codes []string `json:"recovery_codes"`
}

type ProfilePictureId struct {
	Base
	ProfilePictureId string `json:"profile_picture_id"`
}

func NewInvalidCredentials(lang string) Base {
	return New(false, "Invalid credentials", lang)
}

func NewUserNoLongerExists(lang string) Base {
	return New(false, "The user does no longer exist", lang)
}
