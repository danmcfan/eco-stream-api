package models

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID       string
	Username string
	Password string
	IsActive bool
}

type Token struct {
	Token string `json:"token"`
}

type CreateItem struct {
	Name string `json:"name"`
}

type UpdateItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Item struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Count  int    `json:"count"`
	UserID string
}
