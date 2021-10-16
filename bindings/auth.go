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

type VerifyCode struct {
	Email string `json:"email" form:"email"`
	Code  string `json:"code" form:"code"`
}

type PasswordAuth struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type Login struct {
	Email         string `json:"email" form:"email"`
	PasswordToken string `json:"password_token" form:"password_token"`
	TwoFAToken    string `json:"two_fa_token" form:"two_fa_token"`
}

type Password struct {
	Password string `json:"password" form:"password"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" form:"old_password"`
	NewPassword string `json:"new_password" form:"new_password"`
}

type ForgotPassword struct {
	Email        string `json:"email" form:"email"`
	CaptchaToken string `json:"h-captcha-response" form:"h-captcha-response"`
	TwoFAToken   string `json:"two_fa_token" form:"two_fa_token"`
}

type ResetPassword struct {
	Email       string `json:"email" form:"email"`
	Token       string `json:"token" form:"token"`
	NewPassword string `json:"new_password" form:"new_password"`
}

type ChangeEmailRequest struct {
	Password     string `json:"password" form:"password"`
	CaptchaToken string `json:"h-captcha-response" form:"h-captcha-response"`
	NewEmail     string `json:"new_email" form:"new_email"`
}

type ChangeEmail struct {
	Token string `json:"token" form:"token"`
}

type Refresh struct {
	UserId string `json:"user_id" form:"user_id"`
}
