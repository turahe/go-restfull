package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Post represents the core post domain entity
type Post struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Subtitle    string     `json:"subtitle"`
	Description string     `json:"description"`
	IsSticky    bool       `json:"is_sticky"`
	Language    string     `json:"language"`
	Layout      string     `json:"layout"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   uuid.UUID  `json:"updated_by"`
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewPost creates a new post with validation
func NewPost(title, slug, subtitle, description, language, layout string, isSticky bool, publishedAt *time.Time) (*Post, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	if subtitle == "" {
		return nil, errors.New("subtitle is required")
	}
	if description == "" {
		return nil, errors.New("description is required")
	}
	if language == "" {
		return nil, errors.New("language is required")
	}
	if layout == "" {
		return nil, errors.New("layout is required")
	}

	now := time.Now()
	return &Post{
		ID:          uuid.New(),
		Title:       title,
		Slug:        slug,
		Subtitle:    subtitle,
		Description: description,
		IsSticky:    isSticky,
		Language:    language,
		Layout:      layout,
		CreatedAt:   now,
		UpdatedAt:   now,
		PublishedAt: publishedAt,
	}, nil
}

// UpdatePost updates post information
func (p *Post) UpdatePost(title, slug, subtitle, description, language, layout string, isSticky bool, publishedAt *time.Time) error {
	if title != "" {
		p.Title = title
	}
	if slug != "" {
		p.Slug = slug
	}
	if subtitle != "" {
		p.Subtitle = subtitle
	}
	if description != "" {
		p.Description = description
	}
	if language != "" {
		p.Language = language
	}
	if layout != "" {
		p.Layout = layout
	}
	if isSticky != p.IsSticky {
		p.IsSticky = isSticky
	}
	if publishedAt != nil {
		p.PublishedAt = publishedAt
	}
	p.UpdatedAt = time.Now()
	return nil
}

// Publish marks the post as published
func (p *Post) Publish() {
	now := time.Now()
	p.PublishedAt = &now
	p.UpdatedAt = now
}

// Unpublish marks the post as draft
func (p *Post) Unpublish() {
	p.PublishedAt = nil
	p.UpdatedAt = time.Now()
}

// SoftDelete marks the post as deleted
func (p *Post) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}

// IsDeleted checks if the post is soft deleted
func (p *Post) IsDeleted() bool {
	return p.DeletedAt != nil
}

// IsPublished checks if the post is published
func (p *Post) IsPublished() bool {
	return p.PublishedAt != nil
}

// IsDraft checks if the post is in draft status
func (p *Post) IsDraft() bool {
	return p.PublishedAt == nil
}
