package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

type User struct {
	ID        int
	GroupID   int
	Username  string
	Password  string
	Name      string
	CreatedAt time.Time
}

type JWTUser struct {
	ID       int
	Username string
	Password string
	Name     string
	Type     string
}

type Dropdown struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AllItemItem struct {
	ItemID         string  `json:"item_id"`
	ModelID        string  `json:"model_id"`
	ItemCategoryID string  `json:"item_category_id"`
	PageNo         string  `json:"page_no"`
	ItemNo         string  `json:"item_no"`
	ForeignID      string  `json:"foreign_id"`
	ItemName       string  `json:"item_name"`
	Price          float64 `json:"price"`
}
