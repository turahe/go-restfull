// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Tag entity for managing
// content tags and labels with visual customization support.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Tag represents the core tag domain entity that provides content categorization
// and labeling functionality. Tags are used to organize and classify content
// across the system with visual customization through color coding.
//
// The entity includes:
// - Tag identification (name, slug, description)
// - Visual customization (color for UI display)
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for tag preservation
// - Content organization and classification support
type Tag struct {
	ID          uuid.UUID  `json:"id"`                   // Unique identifier for the tag
	Name        string     `json:"name"`                 // Display name of the tag
	Slug        string     `json:"slug"`                 // URL-friendly identifier for the tag
	Description string     `json:"description"`          // Description of the tag's purpose
	Color       string     `json:"color"`                // Color code for visual representation (hex, CSS color, etc.)
	CreatedBy   uuid.UUID  `json:"created_by"`           // ID of user who created this tag
	UpdatedBy   uuid.UUID  `json:"updated_by"`           // ID of user who last updated this tag
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"` // ID of user who deleted this tag (soft delete)
	CreatedAt   time.Time  `json:"created_at"`           // Timestamp when tag was created
	UpdatedAt   time.Time  `json:"updated_at"`           // Timestamp when tag was last updated
	DeletedAt   *time.Time `json:"deleted_at,omitempty"` // Timestamp when tag was soft deleted
}

// NewTag creates a new tag with validation.
// This constructor validates required fields and initializes the tag
// with generated UUID and timestamps.
//
// Parameters:
//   - name: Display name of the tag (required)
//   - slug: URL-friendly identifier (required)
//   - description: Description of the tag's purpose
//   - color: Color code for visual representation
//
// Returns:
//   - *Tag: Pointer to the newly created tag entity
//   - error: Validation error if name or slug is empty
//
// Validation rules:
// - name and slug cannot be empty
// - description and color are optional
//
// Note: Color can be any valid color format (hex, CSS color names, RGB, etc.)
func NewTag(name, slug, description, color string) (*Tag, error) {
	// Validate required fields
	if name == "" {
		return nil, errors.New("name is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	// Create tag with current timestamp
	now := time.Now()
	return &Tag{
		ID:          uuid.New(),  // Generate new unique identifier
		Name:        name,        // Set tag name
		Slug:        slug,        // Set tag slug
		Description: description, // Set tag description
		Color:       color,       // Set tag color
		CreatedAt:   now,         // Set creation timestamp
		UpdatedAt:   now,         // Set initial update timestamp
	}, nil
}

// UpdateTag updates tag information with new values.
// This method allows partial updates - only non-empty values are updated.
// The UpdatedAt timestamp is automatically updated when any field changes.
//
// Parameters:
//   - name: New tag name (optional, only updated if not empty)
//   - slug: New tag slug (optional, only updated if not empty)
//   - description: New tag description (optional, only updated if not empty)
//   - color: New tag color (optional, only updated if not empty)
//
// Returns:
//   - error: Always nil, included for interface consistency
//
// Note: This method automatically updates the UpdatedAt timestamp
func (t *Tag) UpdateTag(name, slug, description, color string) error {
	// Update fields only if new values are provided
	if name != "" {
		t.Name = name
	}
	if slug != "" {
		t.Slug = slug
	}
	if description != "" {
		t.Description = description
	}
	if color != "" {
		t.Color = color
	}

	// Update modification timestamp
	t.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the tag as deleted without removing it from the database.
// This sets the DeletedAt timestamp and updates the UpdatedAt timestamp.
// The tag will be excluded from normal queries but remains accessible
// for audit and recovery purposes.
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (t *Tag) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now // Set deletion timestamp
	t.UpdatedAt = now  // Update modification timestamp
}

// IsDeleted checks if the tag has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
// This method is useful for filtering out deleted tags from queries.
//
// Returns:
//   - bool: true if tag is deleted, false if active
func (t *Tag) IsDeleted() bool {
	return t.DeletedAt != nil
}
