package responses

type Generic struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
