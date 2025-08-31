// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreateRoleRequest represents the request for creating a new role entity.
// This struct defines the required and optional fields for role creation,
// including validation tags for field constraints and business rules.
type CreateRoleRequest struct {
	// Name is the display name for the role (required, 1-255 characters)
	Name string `json:"name" validate:"required,min=1,max=255"`
	// Slug is the URL-friendly identifier for the role (optional, 1-255 characters if provided, auto-generated from name if empty)
	Slug string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	// Description provides additional details about the role (optional, max 1000 characters)
	Description string `json:"description,omitempty" validate:"max=1000"`
}

// UpdateRoleRequest represents the request for updating an existing role entity.
// This struct defines the fields that can be updated for a role,
// including validation tags for field constraints and business rules.
type UpdateRoleRequest struct {
	// Name is the display name for the role (optional, 1-255 characters if provided)
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// Slug is the URL-friendly identifier for the role (optional, 1-255 characters if provided)
	Slug string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	// Description provides additional details about the role (optional, max 1000 characters if provided)
	Description string `json:"description,omitempty" validate:"max=1000"`
}

// generateRoleSlug creates a URL-friendly slug from a given string.
// This function converts the input to lowercase, replaces spaces and special characters with hyphens,
// and removes any non-alphanumeric characters except hyphens.
//
// Parameters:
//   - input: The string to convert to a slug
//
// Returns:
//   - string: The generated slug
func generateRoleSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// If the result is empty, use a default slug
	if slug == "" {
		slug = "role"
	}

	return slug
}

// Validate performs validation on the CreateRoleRequest using the validator package.
// This method checks all field constraints including required fields and length limits.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateRoleRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs validation on the UpdateRoleRequest using the validator package.
// This method checks all field constraints including length limits.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *UpdateRoleRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the CreateRoleRequest to a Role domain entity.
// This method handles optional fields, generates a new UUID for the role entity,
// and sets default values.
//
// Returns:
//   - *entities.Role: The created role entity
func (r *CreateRoleRequest) ToEntity() *entities.Role {
	// Generate slug from name if not provided
	slug := r.Slug
	if slug == "" {
		slug = generateRoleSlug(r.Name)
	}

	// Create and populate the role entity
	role := &entities.Role{
		ID:          uuid.New(),
		Name:        r.Name,
		Slug:        slug,
		Description: r.Description,
		IsActive:    true, // Default to active
	}

	return role
}

// ToEntity transforms the UpdateRoleRequest to update an existing Role domain entity.
// This method updates the role entity with the new values provided in the request.
//
// Parameters:
//   - existingRole: The existing role entity to update
//
// Returns:
//   - *entities.Role: The updated role entity
func (r *UpdateRoleRequest) ToEntity(existingRole *entities.Role) *entities.Role {
	// Update fields if provided, otherwise keep existing values
	if r.Name != "" {
		existingRole.Name = r.Name
	}
	if r.Slug != "" {
		existingRole.Slug = r.Slug
	}
	if r.Description != "" {
		existingRole.Description = r.Description
	}

	return existingRole
}
