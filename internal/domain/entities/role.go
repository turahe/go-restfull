// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Role entity for managing
// user roles and permissions in the role-based access control system.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role represents a role entity in the domain layer that defines
// user permissions and access levels within the system.
//
// The entity includes:
// - Role identification (name, slug, description)
// - Status management (active/inactive)
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for role preservation
// - Integration with RBAC (Role-Based Access Control) system
type Role struct {
	ID          uuid.UUID  `json:"id"`                    // Unique identifier for the role
	Name        string     `json:"name"`                  // Display name of the role
	Slug        string     `json:"slug"`                  // URL-friendly identifier for the role
	Description string     `json:"description,omitempty"` // Optional description of the role's purpose
	IsActive    bool       `json:"is_active"`             // Whether the role is currently active
	CreatedBy   uuid.UUID  `json:"created_by"`            // ID of user who created this role
	UpdatedBy   uuid.UUID  `json:"updated_by"`            // ID of user who last updated this role
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"`  // ID of user who deleted this role (soft delete)
	CreatedAt   time.Time  `json:"created_at"`            // Timestamp when role was created
	UpdatedAt   time.Time  `json:"updated_at"`            // Timestamp when role was last updated
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`  // Timestamp when role was soft deleted
}

// NewRole creates a new role with validation.
// This constructor validates required fields and initializes the role
// with default active status and generated UUID and timestamps.
//
// Parameters:
//   - name: Display name of the role (required)
//   - slug: URL-friendly identifier (required)
//   - description: Optional description of the role's purpose
//
// Returns:
//   - *Role: Pointer to the newly created role entity
//   - error: Validation error if name or slug is empty
//
// Validation rules:
// - name and slug cannot be empty
// - description is optional
//
// Default values:
//   - IsActive: true (role is active by default)
//   - CreatedAt/UpdatedAt: Current timestamp
func NewRole(name, slug, description string) (*Role, error) {
	// Validate required fields
	if name == "" {
		return nil, errors.New("role name is required")
	}
	if slug == "" {
		return nil, errors.New("role slug is required")
	}

	// Create role with current timestamp
	now := time.Now()
	return &Role{
		ID:          uuid.New(),  // Generate new unique identifier
		Name:        name,        // Set role name
		Slug:        slug,        // Set role slug
		Description: description, // Set role description
		IsActive:    true,        // Set as active by default
		CreatedAt:   now,         // Set creation timestamp
		UpdatedAt:   now,         // Set initial update timestamp
	}, nil
}

// UpdateRole updates role information with new values.
// This method allows partial updates - only non-empty values are updated.
// The UpdatedAt timestamp is automatically updated when any field changes.
//
// Parameters:
//   - name: New role name (optional, only updated if not empty)
//   - slug: New role slug (optional, only updated if not empty)
//   - description: New role description (always updated, can be empty)
//
// Returns:
//   - error: Always nil, included for interface consistency
//
// Note: This method automatically updates the UpdatedAt timestamp
func (r *Role) UpdateRole(name, slug, description string) error {
	// Update fields only if new values are provided
	if name != "" {
		r.Name = name
	}
	if slug != "" {
		r.Slug = slug
	}

	// Always update description (can be empty)
	r.Description = description

	// Update modification timestamp
	r.UpdatedAt = time.Now()
	return nil
}

// Activate enables the role.
// This method sets IsActive to true and updates the UpdatedAt timestamp.
// Active roles are typically included in permission checks and user assignments.
//
// Note: This method automatically updates the UpdatedAt timestamp
func (r *Role) Activate() {
	r.IsActive = true        // Enable the role
	r.UpdatedAt = time.Now() // Update modification timestamp
}

// Deactivate disables the role.
// This method sets IsActive to false and updates the UpdatedAt timestamp.
// Inactive roles are typically excluded from permission checks and new user assignments.
//
// Note: This method automatically updates the UpdatedAt timestamp
func (r *Role) Deactivate() {
	r.IsActive = false       // Disable the role
	r.UpdatedAt = time.Now() // Update modification timestamp
}

// SoftDelete marks the role as deleted without removing it from the database.
// This sets the DeletedAt timestamp and updates the UpdatedAt timestamp.
// The role will be excluded from normal queries but remains accessible
// for audit and recovery purposes.
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (r *Role) SoftDelete() {
	now := time.Now()
	r.DeletedAt = &now // Set deletion timestamp
	r.UpdatedAt = now  // Update modification timestamp
}

// IsDeleted checks if the role has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
// This method is useful for filtering out deleted roles from queries.
//
// Returns:
//   - bool: true if role is deleted, false if active
func (r *Role) IsDeleted() bool {
	return r.DeletedAt != nil
}

// IsActiveRole checks if the role is both active and not deleted.
// This method provides a comprehensive check for whether a role
// should be considered available for use in the system.
//
// Returns:
//   - bool: true if role is active and not deleted, false otherwise
//
// Note: Combines IsActive and !IsDeleted checks for complete status validation
func (r *Role) IsActiveRole() bool {
	return r.IsActive && !r.IsDeleted()
}
