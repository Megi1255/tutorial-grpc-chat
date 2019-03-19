package domain

type User struct {
	Token string `json:"user_id"`
	Name  string `json:"name"`
}
