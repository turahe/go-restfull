package model

import (
	"time"
)

type MenuRole struct {
	ID        string    `json:"id"`
	MenuID    string    `json:"menu_id"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
