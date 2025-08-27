// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"errors"
	"regexp"

	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreateTagRequest represents the request for creating a new tag entity.
// This struct defines the required and optional fields for tag creation,
// including name, slug, description, and color. Tags are used for
// categorizing and organizing content throughout the system.
type CreateTagRequest struct {
	// Name is the display name for the tag (required)
	Name string `json:"name"`
	// Slug is the URL-friendly identifier for the tag (required, lowercase, alphanumeric + hyphens only)
	Slug string `json:"slug"`
	// Description provides additional details about the tag (optional)
	Description string `json:"description"`
	// Color is the hex color code for visual representation (optional, must be valid hex format if provided)
	Color string `json:"color"`
}

// Validate performs comprehensive validation on the CreateTagRequest.
// This method checks all required fields and validates the format of
// slug and color fields using regex patterns.
//
// Validation Rules:
// - Name and slug are required
// - Slug must contain only lowercase letters, numbers, and hyphens
// - Color must be a valid 6-digit hex color code if provided
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateTagRequest) Validate() error {
	// Check required fields
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Slug == "" {
		return errors.New("slug is required")
	}

	// Validate slug format (alphanumeric and hyphens only, lowercase)
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !slugRegex.MatchString(r.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate color format (hex color) if provided
	if r.Color != "" {
		colorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
		if !colorRegex.MatchString(r.Color) {
			return errors.New("color must be a valid hex color (e.g., #FF0000)")
		}
	}

	return nil
}

// ToEntity transforms the CreateTagRequest to a Tag domain entity.
// This method creates a new tag with a generated UUID and populates
// all the provided fields for persistence.
//
// Returns:
//   - *entities.Tag: The created tag entity
func (r *CreateTagRequest) ToEntity() *entities.Tag {
	return &entities.Tag{
		ID:          uuid.New(),
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		Color:       r.Color,
	}
}

// UpdateTagRequest represents the request for updating an existing tag entity.
// This struct allows updating tag properties while maintaining data integrity.
// All fields are treated as optional for partial updates.
type UpdateTagRequest struct {
	// Name is the display name for the tag (optional)
	Name string `json:"name"`
	// Slug is the URL-friendly identifier for the tag (optional, lowercase, alphanumeric + hyphens only)
	Slug string `json:"slug"`
	// Description provides additional details about the tag (optional)
	Description string `json:"description"`
	// Color is the hex color code for visual representation (optional, must be valid hex format if provided)
	Color string `json:"color"`
}

// Validate performs validation on the UpdateTagRequest for any provided fields.
// This method validates the format of slug and color fields using regex patterns
// only if they are provided in the update request.
//
// Validation Rules:
// - Name and slug are required if provided
// - Slug must contain only lowercase letters, numbers, and hyphens
// - Color must be a valid 6-digit hex color code if provided
//
// Returns:
//   - error: Validation error if any provided field fails validation, nil if valid
func (r *UpdateTagRequest) Validate() error {
	// Validate name if provided
	if r.Name == "" {
		return errors.New("name is required")
	}
	// Validate slug if provided
	if r.Slug == "" {
		return errors.New("slug is required")
	}

	// Validate slug format (alphanumeric and hyphens only, lowercase)
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !slugRegex.MatchString(r.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate color format (hex color) if provided
	if r.Color != "" {
		colorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
		if !colorRegex.MatchString(r.Color) {
			return errors.New("color must be a valid hex color (e.g., #FF0000)")
		}
	}

	return nil
}

// ToEntity transforms the UpdateTagRequest to update an existing Tag entity.
// This method applies only the provided fields to the existing tag, preserving
// unchanged values. It's designed for partial updates where not all fields are provided.
//
// Parameters:
//   - existingTag: The existing tag entity to update
//
// Returns:
//   - *entities.Tag: The updated tag entity
func (r *UpdateTagRequest) ToEntity(existingTag *entities.Tag) *entities.Tag {
	// Update fields only if provided in the request
	if r.Name != "" {
		existingTag.Name = r.Name
	}
	if r.Slug != "" {
		existingTag.Slug = r.Slug
	}
	if r.Description != "" {
		existingTag.Description = r.Description
	}
	if r.Color != "" {
		existingTag.Color = r.Color
	}

	return existingTag
}

// SearchTagsRequest represents the request for searching and filtering tags.
// This struct supports text-based search with pagination parameters for
// result management and efficient data retrieval.
type SearchTagsRequest struct {
	// Query is the search term to match against tag names and descriptions (optional)
	Query string `json:"query" query:"query"`
	// Limit controls the maximum number of results returned (optional, must be non-negative, max 100)
	Limit int `json:"limit" query:"limit"`
	// Offset controls the number of results to skip for pagination (optional, must be non-negative)
	Offset int `json:"offset" query:"offset"`
}

// Validate performs validation on the SearchTagsRequest using the validator package.
// This method checks field constraints for pagination parameters to ensure
// reasonable limits and prevent excessive resource usage.
//
// Validation Rules:
// - Limit must be non-negative and cannot exceed 100
// - Offset must be non-negative
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *SearchTagsRequest) Validate() error {
	// Validate pagination parameters
	if r.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if r.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	// Prevent excessive resource usage with reasonable limit
	if r.Limit > 100 {
		return errors.New("limit cannot exceed 100")
	}
	return nil
}
