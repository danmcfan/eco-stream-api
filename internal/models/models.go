package models

type CreateUser struct {
	Username string `json:"username"`
}

type UpdateUser struct {
	Username string `json:"username"`
	IsActive bool   `json:"isActive"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsActive bool   `json:"isActive"`
}
