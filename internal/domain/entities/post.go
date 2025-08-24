// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Post entity for managing
// blog posts and articles with publishing workflow and content management.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Post represents the core post domain entity that manages blog posts,
// articles, and other content with publishing workflow support.
//
// The entity includes:
// - Content management (title, subtitle, description, slug)
// - Publishing workflow (draft, published states)
// - Content presentation (layout, language, sticky status)
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for content preservation
type Post struct {
	ID          uuid.UUID  `json:"id"`                     // Unique identifier for the post
	Title       string     `json:"title"`                  // Main title of the post
	Slug        string     `json:"slug"`                   // URL-friendly identifier for the post
	Subtitle    string     `json:"subtitle"`               // Secondary title or subtitle
	Description string     `json:"description"`            // Brief description or summary of the post
	Type        string     `json:"type"`                   // Type of post (e.g., "post", "page", "article")
	IsSticky    bool       `json:"is_sticky"`              // Whether the post should be pinned/sticky
	Language    string     `json:"language"`               // Language code for the post content
	Layout      string     `json:"layout"`                 // Layout template identifier for rendering
	Content     string     `json:"content"`                // insert to table contents
	PublishedAt *time.Time `json:"published_at,omitempty"` // Timestamp when post was published (nil for drafts)
	CreatedBy   uuid.UUID  `json:"created_by"`             // ID of user who created this post
	UpdatedBy   uuid.UUID  `json:"updated_by"`             // ID of user who last updated this post
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"`   // ID of user who deleted this post (soft delete)
	CreatedAt   time.Time  `json:"created_at"`             // Timestamp when post was created
	UpdatedAt   time.Time  `json:"updated_at"`             // Timestamp when post was last updated
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`   // Timestamp when post was soft deleted
}

// NewPost creates a new post with validation.
// This constructor validates required fields and initializes the post
// with generated UUID and timestamps.
//
// Parameters:
//   - title: Main title of the post (required)
//   - slug: URL-friendly identifier (required)
//   - subtitle: Secondary title or subtitle (required)
//   - description: Brief description or summary (required)
//   - type: Type of post (e.g., "post", "page", "article") (required)
//   - language: Language code for content (required)
//   - layout: Layout template identifier (required)
//   - content: Content body of the post (required)
//   - createdBy: UUID of the user creating the post (required)
//   - isSticky: Whether post should be pinned/sticky
//   - publishedAt: Optional timestamp for immediate publishing (nil for drafts)
//
// Returns:
//   - *Post: Pointer to the newly created post entity
//   - error: Validation error if any required field is empty
//
// Validation rules:
// - title, slug, subtitle, description, type, language, content, layout, and createdBy cannot be empty
func NewPost(title, slug, subtitle, description, postType, language, layout, content string, createdBy uuid.UUID, isSticky bool, publishedAt *time.Time) (*Post, error) {
	// Validate required fields
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
	if postType == "" {
		return nil, errors.New("type is required")
	}
	if language == "" {
		return nil, errors.New("language is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}
	if layout == "" {
		return nil, errors.New("layout is required")
	}
	if createdBy == uuid.Nil {
		return nil, errors.New("created_by is required")
	}

	// Create post with current timestamp
	now := time.Now()
	return &Post{
		ID:          uuid.New(),  // Generate new unique identifier
		Title:       title,       // Set post title
		Slug:        slug,        // Set post slug
		Subtitle:    subtitle,    // Set post subtitle
		Description: description, // Set post description
		Type:        postType,    // Set post type
		IsSticky:    isSticky,    // Set sticky status
		Language:    language,    // Set content language
		Layout:      layout,      // Set layout template
		Content:     content,     // Set content
		CreatedBy:   createdBy,   // Set creator ID
		UpdatedBy:   createdBy,   // Initially same as creator
		CreatedAt:   now,         // Set creation timestamp
		UpdatedAt:   now,         // Set initial update timestamp
		PublishedAt: publishedAt, // Set publication timestamp (may be nil for drafts)
	}, nil
}

// UpdatePost updates post information with new values.
// This method allows partial updates - only non-empty values are updated.
// The UpdatedAt timestamp is automatically updated when any field changes.
//
// Parameters:
//   - title: New post title (optional, only updated if not empty)
//   - slug: New post slug (optional, only updated if not empty)
//   - subtitle: New post subtitle (optional, only updated if not empty)
//   - description: New post description (optional, only updated if not empty)
//   - type: New post type (optional, only updated if not empty)
//   - language: New content language (optional, only updated if not empty)
//   - layout: New layout template (optional, only updated if not empty)
//   - isSticky: New sticky status
//   - publishedAt: New publication timestamp (optional, only updated if not nil)
//
// Returns:
//   - error: Always nil, included for interface consistency
//
// Note: This method automatically updates the UpdatedAt timestamp
func (p *Post) UpdatePost(title, slug, subtitle, description, postType, language, layout string, isSticky bool, publishedAt *time.Time) error {
	// Update fields only if new values are provided
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
	if postType != "" {
		p.Type = postType
	}
	if language != "" {
		p.Language = language
	}
	if layout != "" {
		p.Layout = layout
	}

	// Update boolean and pointer fields
	if isSticky != p.IsSticky {
		p.IsSticky = isSticky
	}
	if publishedAt != nil {
		p.PublishedAt = publishedAt
	}

	// Update modification timestamp
	p.UpdatedAt = time.Now()
	return nil
}

// Publish marks the post as published.
// This method sets the PublishedAt timestamp to the current time
// and updates the UpdatedAt timestamp. Once published, the post
// becomes visible to readers and appears in published content queries.
//
// Note: This method automatically updates both PublishedAt and UpdatedAt timestamps
func (p *Post) Publish() {
	now := time.Now()
	p.PublishedAt = &now // Set publication timestamp
	p.UpdatedAt = now    // Update modification timestamp
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
