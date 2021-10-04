package bindings

type Register struct {
	Name         string `json:"name" form:"name"`
	Email        string `json:"email" form:"email"`
	Password     string `json:"password" form:"password"`
	CaptchaToken string `json:"h-captcha-response" form:"h-captcha-response"`
}

type ConfirmEmail struct {
	Email string `json:"email" form:"email"`
	Code  string `json:"code" form:"code"`
}

type Activate2FAOTP struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type VerifyOTPCode struct {
	Email      string `json:"email" form:"email"`
	OTPCode    string `json:"otp_code" form:"otp_code"`
	LoginToken string `json:"login_token" form:"login_token"`
}

type Login struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}
