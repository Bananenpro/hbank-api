package responses

type Token struct {
	Base
	Token string `json:"token"`
}

type RecoveryCodes struct {
	Base
	Codes []string `json:"recoveryCodes"`
}

func NewInvalidCredentials(lang string) Base {
	return New(false, "Invalid credentials", lang)
}

func NewUserNoLongerExists(lang string) Base {
	return New(false, "The user does no longer exist", lang)
}
