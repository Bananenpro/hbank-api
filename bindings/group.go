package bindings

type CreateGroup struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	OnlyAdmin   bool   `json:"only_admin" form:"only_admin"`
}
