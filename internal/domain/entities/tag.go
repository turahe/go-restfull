package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Tag represents the core tag domain entity
type Tag struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewTag creates a new tag with validation
func NewTag(name, slug, description, color string) (*Tag, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	now := time.Now()
	return &Tag{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: description,
		Color:       color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateTag updates tag information
func (t *Tag) UpdateTag(name, slug, description, color string) error {
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
	t.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the tag as deleted
func (t *Tag) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now
	t.UpdatedAt = now
}

// IsDeleted checks if the tag is deleted
func (t *Tag) IsDeleted() bool {
	return t.DeletedAt != nil
}
