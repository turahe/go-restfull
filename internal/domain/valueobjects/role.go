package valueobjects

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Role represents a role value object
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewRole creates a new role value object
func NewRole(id uuid.UUID, name, description string, createdAt time.Time) (Role, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if id == uuid.Nil {
		return Role{}, errors.New("role ID cannot be nil")
	}

	if name == "" {
		return Role{}, errors.New("role name cannot be empty")
	}

	if len(name) > 50 {
		return Role{}, errors.New("role name cannot exceed 50 characters")
	}

	if len(description) > 255 {
		return Role{}, errors.New("role description cannot exceed 255 characters")
	}

	return Role{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

// Equals checks if two roles are equal
func (r Role) Equals(other Role) bool {
	return r.ID == other.ID
}

// String returns the string representation of the role
func (r Role) String() string {
	return r.Name
}