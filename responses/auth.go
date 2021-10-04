package responses

type RegisterSuccess struct {
	Generic
	UserId    string `json:"user_id"`
	UserEmail string `json:"user_email"`
}

type RegisterInvalid struct {
	Generic
	MinNameLength     int `json:"min_name_length"`
	MinPasswordLength int `json:"min_password_length"`
	MaxNameLength     int `json:"max_name_length"`
	MaxPasswordLength int `json:"max_password_length"`
}
