package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleListResponse represents a list of roles with pagination
type RoleListResponse struct {
	Roles []RoleResponse `json:"roles"`
	Total int64          `json:"total"`
	Limit int            `json:"limit"`
	Page  int            `json:"page"`
}

// NewRoleResponse creates a new RoleResponse from role entity
func NewRoleResponse(role *entities.Role) *RoleResponse {
	return &RoleResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

// NewRoleListResponse creates a new RoleListResponse from role entities
func NewRoleListResponse(roles []*entities.Role, total int64, limit, page int) *RoleListResponse {
	roleResponses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = *NewRoleResponse(role)
	}

	return &RoleListResponse{
		Roles: roleResponses,
		Total: total,
		Limit: limit,
		Page:  page,
	}
}
