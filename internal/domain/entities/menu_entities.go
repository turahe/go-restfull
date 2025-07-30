package entities

import (
	"time"

	"github.com/google/uuid"
)

// MenuEntities represents the many-to-many relationship between menus and roles
type MenuEntities struct {
	ID         uuid.UUID `json:"id"`
	MenuID     uuid.UUID `json:"menu_id"`
	EntityID   uuid.UUID `json:"entity_id"`
	EntityType string    `json:"entity_type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NewMenuRole creates a new menu-role relationship
func NewMenuRole(menuID, roleID uuid.UUID, entityType string) *MenuEntities {
	return &MenuEntities{
		ID:         uuid.New(),
		MenuID:     menuID,
		EntityID:   roleID,
		EntityType: entityType,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
