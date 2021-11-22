package bindings

type Register struct {
	Name         string `json:"name" form:"name"`
	Email        string `json:"email" form:"email"`
	Password     string `json:"password" form:"password"`
	CaptchaToken string `json:"hCaptchaResponse" form:"h-captcha-response"`
}

type ConfirmEmail struct {
	Email string `json:"email" form:"email"`
	Code  string `json:"code" form:"code"`
}

type EmailPassword struct {
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
	PasswordToken string `json:"passwordToken" form:"passwordToken"`
	TwoFAToken    string `json:"twoFAToken" form:"twoFAToken"`
}

type Password struct {
	Password string `json:"password" form:"password"`
}

type ChangePassword struct {
	OldPassword string `json:"oldPassword" form:"oldPassword"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}

type ForgotPassword struct {
	Email        string `json:"email" form:"email"`
	CaptchaToken string `json:"hCaptchaResponse" form:"h-captcha-response"`
	TwoFAToken   string `json:"twoFAToken" form:"twoFAToken"`
}

type ResetPassword struct {
	Email       string `json:"email" form:"email"`
	Token       string `json:"token" form:"token"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}

type ChangeEmailRequest struct {
	Password     string `json:"password" form:"password"`
	CaptchaToken string `json:"hCaptchaResponse" form:"hCaptchaResponse"`
	NewEmail     string `json:"newEmail" form:"newEmail"`
}

type ChangeEmail struct {
	Token string `json:"token" form:"token"`
}

type Refresh struct {
	UserId string `json:"userId" form:"userId"`
}
