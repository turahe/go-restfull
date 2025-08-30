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

// CreateTaxonomyRequest represents the request for creating a new taxonomy entity.
// This struct defines the required and optional fields for taxonomy creation,
// including validation tags for field constraints and business rules.
// The request supports hierarchical taxonomy structures through parent_id.
type CreateTaxonomyRequest struct {
	// Name is the display name for the taxonomy (required, 1-255 characters)
	Name string `json:"name" validate:"required,min=1,max=255"`
	// Slug is the URL-friendly identifier for the taxonomy (optional, 1-255 characters if provided, auto-generated from name if empty)
	Slug string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	// Code is a unique identifier for the taxonomy (optional, max 50 characters)
	Code string `json:"code,omitempty" validate:"max=50"`
	// Description provides additional details about the taxonomy (optional, max 1000 characters)
	Description string `json:"description,omitempty" validate:"max=1000"`
	// ParentID is the UUID of the parent taxonomy for hierarchical structures (optional, must be valid UUID if provided)
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
}

// generateTaxonomySlug creates a URL-friendly slug from a given string.
// This function converts the input to lowercase, replaces spaces and special characters with hyphens,
// and removes any non-alphanumeric characters except hyphens.
//
// Parameters:
//   - input: The string to convert to a slug
//
// Returns:
//   - string: The generated slug
func generateTaxonomySlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// If the result is empty, use a default slug
	if slug == "" {
		slug = "taxonomy"
	}

	return slug
}

// Validate performs validation on the CreateTaxonomyRequest using the validator package.
// This method checks all field constraints including required fields, length limits,
// and UUID format validation for the parent_id field.
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateTaxonomyRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the CreateTaxonomyRequest to a Taxonomy domain entity.
// This method parses the parent_id string to UUID if provided, handles optional fields,
// and generates a new UUID for the taxonomy entity.
//
// Returns:
//   - *entities.Taxonomy: The created taxonomy entity
//   - error: Any error that occurred during transformation (e.g., UUID parsing)
func (r *CreateTaxonomyRequest) ToEntity() (*entities.Taxonomy, error) {
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
		slug = generateTaxonomySlug(r.Name)
	}

	// Create and populate the taxonomy entity
	taxonomy := &entities.Taxonomy{
		ID:          uuid.New(),
		Name:        r.Name,
		Slug:        slug,
		Code:        r.Code,
		Description: r.Description,
		ParentID:    parentID,
	}

	return taxonomy, nil
}

// UpdateTaxonomyRequest represents the request for updating an existing taxonomy entity.
// This struct uses omitempty tags to make all fields optional, allowing partial updates.
// Only provided fields will be updated in the existing taxonomy entity.
type UpdateTaxonomyRequest struct {
	// Name is the display name for the taxonomy (optional, 1-255 characters if provided)
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// Slug is the URL-friendly identifier for the taxonomy (optional, 1-255 characters if provided)
	Slug string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	// Code is a unique identifier for the taxonomy (optional, max 50 characters if provided)
	Code string `json:"code,omitempty" validate:"max=50"`
	// Description provides additional details about the taxonomy (optional, max 1000 characters if provided)
	Description string `json:"description,omitempty" validate:"max=1000"`
	// ParentID is the UUID of the parent taxonomy for hierarchical structures (optional, must be valid UUID if provided)
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
}

// Validate performs validation on the UpdateTaxonomyRequest using the validator package.
// This method checks field constraints for any provided fields while allowing
// all fields to be optional for partial updates.
//
// Returns:
//   - error: Validation error if any provided field fails validation, nil if valid
func (r *UpdateTaxonomyRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ToEntity transforms the UpdateTaxonomyRequest to update an existing Taxonomy entity.
// This method applies only the provided fields to the existing taxonomy, preserving
// unchanged values. It's designed for partial updates where not all fields are provided.
// The method handles UUID parsing for the parent_id field when provided.
//
// Parameters:
//   - existingTaxonomy: The existing taxonomy entity to update
//
// Returns:
//   - *entities.Taxonomy: The updated taxonomy entity
//   - error: Any error that occurred during transformation (e.g., UUID parsing)
func (r *UpdateTaxonomyRequest) ToEntity(existingTaxonomy *entities.Taxonomy) (*entities.Taxonomy, error) {
	// Update fields only if provided in the request
	if r.Name != "" {
		existingTaxonomy.Name = r.Name
		// If name is updated but slug is not provided, regenerate slug from new name
		if r.Slug == "" {
			existingTaxonomy.Slug = generateTaxonomySlug(r.Name)
		}
	}
	if r.Slug != "" {
		existingTaxonomy.Slug = r.Slug
	}
	if r.Code != "" {
		existingTaxonomy.Code = r.Code
	}
	if r.Description != "" {
		existingTaxonomy.Description = r.Description
	}
	if r.ParentID != "" {
		// Parse the new parent ID string to UUID
		parentID, err := uuid.Parse(r.ParentID)
		if err != nil {
			return nil, err
		}
		existingTaxonomy.ParentID = &parentID
	}

	return existingTaxonomy, nil
}
