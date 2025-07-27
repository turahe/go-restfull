package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role represents a role entity in the domain layer
type Role struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewRole creates a new role with validation
func NewRole(name, slug, description string) (*Role, error) {
	if name == "" {
		return nil, errors.New("role name is required")
	}
	if slug == "" {
		return nil, errors.New("role slug is required")
	}

	now := time.Now()
	return &Role{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateRole updates role information
func (r *Role) UpdateRole(name, slug, description string) error {
	if name != "" {
		r.Name = name
	}
	if slug != "" {
		r.Slug = slug
	}
	r.Description = description
	r.UpdatedAt = time.Now()
	return nil
}

// Activate marks the role as active
func (r *Role) Activate() {
	r.IsActive = true
	r.UpdatedAt = time.Now()
}

// Deactivate marks the role as inactive
func (r *Role) Deactivate() {
	r.IsActive = false
	r.UpdatedAt = time.Now()
}

// SoftDelete marks the role as deleted
func (r *Role) SoftDelete() {
	now := time.Now()
	r.DeletedAt = &now
	r.UpdatedAt = now
}

// IsDeleted checks if the role is soft deleted
func (r *Role) IsDeleted() bool {
	return r.DeletedAt != nil
}

// IsActive checks if the role is active
func (r *Role) IsActiveRole() bool {
	return r.IsActive && !r.IsDeleted()
}
