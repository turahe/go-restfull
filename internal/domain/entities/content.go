package entities

import (
	"time"

	"github.com/google/uuid"
)

// Content represents a content entity in the domain layer
type Content struct {
	ID          uuid.UUID  `json:"id"`
	ModelType   string     `json:"model_type"`
	ModelID     uuid.UUID  `json:"model_id"`
	ContentRaw  string     `json:"content_raw"`
	ContentHTML string     `json:"content_html"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   uuid.UUID  `json:"updated_by"`
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewContent creates a new content entity
func NewContent(modelType string, modelID uuid.UUID, contentRaw, contentHTML string, createdBy uuid.UUID) *Content {
	now := time.Now()
	return &Content{
		ID:          uuid.New(),
		ModelType:   modelType,
		ModelID:     modelID,
		ContentRaw:  contentRaw,
		ContentHTML: contentHTML,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsDeleted checks if the content is soft deleted
func (c *Content) IsDeleted() bool {
	return c.DeletedAt != nil
}

// MarkAsDeleted marks the content as deleted
func (c *Content) MarkAsDeleted(deletedBy uuid.UUID) {
	now := time.Now()
	c.DeletedBy = &deletedBy
	c.DeletedAt = &now
	c.UpdatedAt = now
}

// Restore restores the content from soft delete
func (c *Content) Restore(updatedBy uuid.UUID) {
	c.DeletedBy = nil
	c.DeletedAt = nil
	c.UpdatedBy = updatedBy
	c.UpdatedAt = time.Now()
}

// UpdateContent updates the content fields
func (c *Content) UpdateContent(contentRaw, contentHTML string, updatedBy uuid.UUID) {
	c.ContentRaw = contentRaw
	c.ContentHTML = contentHTML
	c.UpdatedBy = updatedBy
	c.UpdatedAt = time.Now()
}
