// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"errors"
	"regexp"
	"strings"
	"time"

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
	// Slug is the URL-friendly identifier for the tag (optional, auto-generated from name if not provided, lowercase, alphanumeric + hyphens only)
	Slug string `json:"slug,omitempty"`
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
// - Name is required
// - Slug is optional (auto-generated from name if not provided)
// - If slug is provided, it must contain only lowercase letters, numbers, and hyphens
// - Color must be a valid 6-digit hex color code if provided
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateTagRequest) Validate() error {
	// Check required fields
	if r.Name == "" {
		return errors.New("name is required")
	}
	// Slug is optional - will be auto-generated from name if not provided

	// Validate slug format if provided (alphanumeric and hyphens only, lowercase)
	if r.Slug != "" {
		slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
		if !slugRegex.MatchString(r.Slug) {
			return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
		}
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

// generateTagSlug creates a URL-friendly slug from a given string.
// This function converts the input to lowercase, replaces spaces and special characters with hyphens,
// and removes any non-alphanumeric characters except hyphens.
//
// Parameters:
//   - input: The string to convert to a slug
//
// Returns:
//   - string: The generated slug
func generateTagSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// If the result is empty, use a default slug
	if slug == "" {
		slug = "tag"
	}

	return slug
}

// ToEntity transforms the CreateTagRequest to a Tag domain entity.
// This method creates a new tag with a generated UUID and populates
// all the provided fields for persistence.
//
// Returns:
//   - *entities.Tag: The created tag entity
func (r *CreateTagRequest) ToEntity() *entities.Tag {
	// Generate slug from name if not provided
	slug := r.Slug
	if slug == "" {
		slug = generateTagSlug(r.Name)
	}

	now := time.Now()
	return &entities.Tag{
		ID:          uuid.New(),
		Name:        r.Name,
		Slug:        slug,
		Description: r.Description,
		Color:       r.Color,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   uuid.Nil, // Will be set by service layer
		UpdatedBy:   uuid.Nil, // Will be set by service layer
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
// - Name is optional (if provided, slug will be auto-generated if not provided)
// - Slug is optional (auto-generated from name if name is provided but slug is not)
// - If slug is provided, it must contain only lowercase letters, numbers, and hyphens
// - Color must be a valid 6-digit hex color code if provided
//
// Returns:
//   - error: Validation error if any provided field fails validation, nil if valid
func (r *UpdateTagRequest) Validate() error {
	// Name is optional in updates
	// Slug is optional - will be auto-generated from name if name is provided but slug is not

	// Validate slug format if provided (alphanumeric and hyphens only, lowercase)
	if r.Slug != "" {
		slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
		if !slugRegex.MatchString(r.Slug) {
			return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
		}
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
		// If name is updated but slug is not provided, regenerate slug from new name
		if r.Slug == "" {
			existingTag.Slug = generateTagSlug(r.Name)
		}
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

	// Always update the updated_at timestamp
	existingTag.UpdatedAt = time.Now()

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
