// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Content entity for managing
// content data with polymorphic relationships and version control.
package entities

import (
	"time"

	"github.com/google/uuid"
)

// Content represents a content entity in the domain layer that stores
// both raw and HTML-formatted content for various model types.
//
// The entity supports:
// - Polymorphic relationships through model_type and model_id fields
// - Dual content storage (raw and HTML formats)
// - Soft delete functionality for content preservation
// - Audit trail with creation, update, and deletion tracking
// - Content versioning through update timestamps
type Content struct {
	ID          uuid.UUID  `json:"id"`                   // Unique identifier for the content
	ModelType   string     `json:"model_type"`           // Type of entity this content belongs to (e.g., "post", "page")
	ModelID     uuid.UUID  `json:"model_id"`             // ID of the entity this content belongs to
	ContentRaw  string     `json:"content_raw"`          // Raw, unformatted content (markdown, plain text, etc.)
	ContentHTML string     `json:"content_html"`         // HTML-formatted version of the content for display
	CreatedBy   uuid.UUID  `json:"created_by"`           // ID of user who created this content
	UpdatedBy   uuid.UUID  `json:"updated_by"`           // ID of user who last updated this content
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"` // ID of user who deleted this content (soft delete)
	DeletedAt   *time.Time `json:"deleted_at,omitempty"` // Timestamp when content was soft deleted
	CreatedAt   time.Time  `json:"created_at"`           // Timestamp when content was created
	UpdatedAt   time.Time  `json:"updated_at"`           // Timestamp when content was last updated
}

// NewContent creates a new content entity with the provided details.
// This constructor initializes required fields and sets default values
// for timestamps and generates a new UUID for the content.
//
// Parameters:
//   - modelType: Type of entity this content belongs to (e.g., "post", "page")
//   - modelID: UUID of the entity this content belongs to
//   - contentRaw: Raw, unformatted content (markdown, plain text, etc.)
//   - contentHTML: HTML-formatted version of the content for display
//   - createdBy: UUID of the user creating this content
//
// Returns:
//   - *Content: Pointer to the newly created content entity
func NewContent(modelType string, modelID uuid.UUID, contentRaw, contentHTML string, createdBy uuid.UUID) *Content {
	now := time.Now()
	return &Content{
		ID:          uuid.New(),  // Generate new unique identifier
		ModelType:   modelType,   // Set the model type
		ModelID:     modelID,     // Set the model ID
		ContentRaw:  contentRaw,  // Set raw content
		ContentHTML: contentHTML, // Set HTML content
		CreatedBy:   createdBy,   // Set creator ID
		UpdatedBy:   createdBy,   // Initially, creator is also updater
		CreatedAt:   now,         // Set creation timestamp
		UpdatedAt:   now,         // Set initial update timestamp
	}
}

// IsDeleted checks if the content has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
// This method is useful for filtering out deleted content from queries.
//
// Returns:
//   - bool: true if content is deleted, false if active
func (c *Content) IsDeleted() bool {
	return c.DeletedAt != nil
}

// MarkAsDeleted marks the content as deleted without removing it from the database.
// This sets the DeletedAt timestamp, records who deleted it, and updates
// the UpdatedAt timestamp. The content will be excluded from normal queries
// but remains accessible for audit and recovery purposes.
//
// Parameters:
//   - deletedBy: UUID of the user performing the deletion
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (c *Content) MarkAsDeleted(deletedBy uuid.UUID) {
	now := time.Now()
	c.DeletedBy = &deletedBy // Record who deleted the content
	c.DeletedAt = &now       // Set deletion timestamp
	c.UpdatedAt = now        // Update modification timestamp
}

// Restore restores the content from soft delete status.
// This clears the DeletedAt and DeletedBy fields, making the content
// visible again in normal queries. The UpdatedAt timestamp is also updated.
//
// Parameters:
//   - updatedBy: UUID of the user restoring the content
//
// Note: This method automatically updates the UpdatedAt timestamp
func (c *Content) Restore(updatedBy uuid.UUID) {
	c.DeletedBy = nil        // Clear deletion user reference
	c.DeletedAt = nil        // Clear deletion timestamp
	c.UpdatedBy = updatedBy  // Set restoration user
	c.UpdatedAt = time.Now() // Update modification timestamp
}

// UpdateContent updates the content fields with new values.
// This method modifies both raw and HTML content and automatically updates
// the UpdatedAt timestamp and UpdatedBy field to reflect the change.
//
// Parameters:
//   - contentRaw: New raw, unformatted content
//   - contentHTML: New HTML-formatted content
//   - updatedBy: UUID of the user updating the content
//
// Note: This method automatically updates the UpdatedAt timestamp and UpdatedBy field
func (c *Content) UpdateContent(contentRaw, contentHTML string, updatedBy uuid.UUID) {
	c.ContentRaw = contentRaw   // Update raw content
	c.ContentHTML = contentHTML // Update HTML content
	c.UpdatedBy = updatedBy     // Set updater ID
	c.UpdatedAt = time.Now()    // Update modification timestamp
}
