package entities

import (
	"time"

	"github.com/google/uuid"
)

// MenuRole represents the many-to-many relationship between menus and roles
type MenuRole struct {
	ID        uuid.UUID `json:"id"`
	MenuID    uuid.UUID `json:"menu_id"`
	RoleID    uuid.UUID `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewMenuRole creates a new menu-role relationship
func NewMenuRole(menuID, roleID uuid.UUID) *MenuRole {
	return &MenuRole{
		ID:        uuid.New(),
		MenuID:    menuID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
