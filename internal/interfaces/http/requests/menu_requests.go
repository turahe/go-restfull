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

// CreateMenuRequest represents the request for creating a new menu entity.
// This struct defines the required and optional fields for menu creation,
// including validation tags for field constraints and business rules.
// The request supports hierarchical menu structures through parent_id.
type CreateMenuRequest struct {
	// Name is the display name for the menu (required, 1-255 characters)
	Name string `json:"name" validate:"required,min=1,max=255"`
	// Slug is the URL-friendly identifier for the menu (optional, 1-255 characters if provided, auto-generated from name if empty)
	Slug string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	// Description provides additional details about the menu (optional, max 1000 characters)
	Description string `json:"description,omitempty" validate:"max=1000"`
	// URL is the destination URL for the menu item (optional, max 500 characters if provided)
	URL string `json:"url,omitempty" validate:"max=500"`
	// Icon is the icon identifier for the menu item (optional, max 100 characters)
	Icon string `json:"icon,omitempty" validate:"max=100"`
	// RecordOrdering determines the display order of menu items (optional, must be >= 0 if provided)
	RecordOrdering *uint64 `json:"record_ordering,omitempty" validate:"omitempty,gte=0"`
	// ParentID is the UUID of the parent menu for hierarchical structures (optional, must be valid UUID if provided)
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
	// Target specifies how the menu link should open (optional, must be valid target if provided)
	Target string `json:"target,omitempty" validate:"omitempty,oneof=_self _blank _parent _top"`
}

// UpdateMenuRequest represents the request for updating an existing menu entity.
// This struct defines the fields that can be updated for a menu,
// including validation tags for field constraints and business rules.
type UpdateMenuRequest struct {
	// Name is the display name for the menu (optional, 1-255 characters if provided)
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// Slug is the URL-friendly identifier for the menu (optional, 1-255 characters if provided)
	Slug string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	// Description provides additional details about the menu (optional, max 1000 characters if provided)
	Description string `json:"description,omitempty" validate:"max=1000"`
	// URL is the destination URL for the menu item (optional, max 500 characters if provided)
	URL string `json:"url,omitempty" validate:"max=500"`
	// Icon is the icon identifier for the menu item (optional, max 100 characters if provided)
	Icon string `json:"icon,omitempty" validate:"max=100"`
	// RecordOrdering determines the display order of menu items (optional, must be >= 0 if provided)
	RecordOrdering *uint64 `json:"record_ordering,omitempty" validate:"omitempty,gte=0"`
	// ParentID is the UUID of the parent menu for hierarchical structures (optional, must be valid UUID if provided)
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
	// Target specifies how the menu link should open (optional, must be valid target if provided)
	Target string `json:"target,omitempty" validate:"omitempty,oneof=_self _blank _parent _top"`
}

// generateMenuSlug creates a URL-friendly slug from a given string.
// This function converts the input to lowercase, replaces spaces and special characters with hyphens,
// and removes any non-alphanumeric characters except hyphens.
//
// Parameters:
//   - input: The string to convert to a slug
//
// Returns:
//   - string: The generated slug
func generateMenuSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// If the result is empty, use a default slug
	if slug == "" {
		slug = "menu"
	}

	return slug
}

// Validate performs validation on the CreateMenuRequest using the validator package.
// This method checks all field constraints including required fields, length limits,
// UUID format validation for the parent_id field, and target validation.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateMenuRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs validation on the UpdateMenuRequest using the validator package.
// This method checks all field constraints including length limits,
// UUID format validation for the parent_id field, and target validation.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *UpdateMenuRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the CreateMenuRequest to a Menu domain entity.
// This method parses the parent_id string to UUID if provided, handles optional fields,
// generates a new UUID for the menu entity, and sets default values.
//
// Returns:
//   - *entities.Menu: The created menu entity
//   - error: Any error that occurred during transformation (e.g., UUID parsing)
func (r *CreateMenuRequest) ToEntity() (*entities.Menu, error) {
	// Parse parent_id string to UUID if provided
	var parentID *uuid.UUID
	if r.ParentID != "" {
		parsedID, err := uuid.Parse(r.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &parsedID
	}

	// Generate slug from name if not provided
	slug := r.Slug
	if slug == "" {
		slug = generateMenuSlug(r.Name)
	}

	// Set default record ordering if not provided
	recordOrdering := uint64(0)
	if r.RecordOrdering != nil {
		recordOrdering = *r.RecordOrdering
	}

	// Set default target if not provided
	target := "_self"
	if r.Target != "" {
		target = r.Target
	}

	// Create and populate the menu entity
	menu := &entities.Menu{
		ID:             uuid.New(),
		Name:           r.Name,
		Slug:           slug,
		Description:    r.Description,
		URL:            r.URL,
		Icon:           r.Icon,
		RecordOrdering: &recordOrdering,
		ParentID:       parentID,
		Target:         target,
	}

	return menu, nil
}

// ToEntity transforms the UpdateMenuRequest to a Menu domain entity for updates.
// This method parses the parent_id string to UUID if provided and handles optional fields.
// The existing menu entity is updated with the new values.
//
// Parameters:
//   - existingMenu: The existing menu entity to update
//
// Returns:
//   - *entities.Menu: The updated menu entity
//   - error: Any error that occurred during transformation (e.g., UUID parsing)
func (r *UpdateMenuRequest) ToEntity(existingMenu *entities.Menu) (*entities.Menu, error) {
	// Parse parent_id string to UUID if provided
	var parentID *uuid.UUID
	if r.ParentID != "" {
		parsedID, err := uuid.Parse(r.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &parsedID
	}

	// Update fields if provided, otherwise keep existing values
	if r.Name != "" {
		existingMenu.Name = r.Name
	}
	if r.Slug != "" {
		existingMenu.Slug = r.Slug
	}
	if r.Description != "" {
		existingMenu.Description = r.Description
	}
	if r.URL != "" {
		existingMenu.URL = r.URL
	}
	if r.Icon != "" {
		existingMenu.Icon = r.Icon
	}
	if r.RecordOrdering != nil {
		existingMenu.RecordOrdering = r.RecordOrdering
	}
	if r.ParentID != "" {
		existingMenu.ParentID = parentID
	}
	if r.Target != "" {
		existingMenu.Target = r.Target
	}

	return existingMenu, nil
}
