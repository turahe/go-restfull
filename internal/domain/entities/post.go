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
	Content     string     `json:"content"`
	Slug        string     `json:"slug"`
	Status      string     `json:"status"`
	AuthorID    uuid.UUID  `json:"author_id"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewPost creates a new post with validation
func NewPost(title, content, slug, status string, authorID uuid.UUID) (*Post, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	if status == "" {
		return nil, errors.New("status is required")
	}
	if authorID == uuid.Nil {
		return nil, errors.New("author_id is required")
	}

	now := time.Now()
	return &Post{
		ID:        uuid.New(),
		Title:     title,
		Content:   content,
		Slug:      slug,
		Status:    status,
		AuthorID:  authorID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdatePost updates post information
func (p *Post) UpdatePost(title, content, slug, status string) error {
	if title != "" {
		p.Title = title
	}
	if content != "" {
		p.Content = content
	}
	if slug != "" {
		p.Slug = slug
	}
	if status != "" {
		p.Status = status
	}
	p.UpdatedAt = time.Now()
	return nil
}

// Publish marks the post as published
func (p *Post) Publish() {
	now := time.Now()
	p.Status = "published"
	p.PublishedAt = &now
	p.UpdatedAt = now
}

// Unpublish marks the post as draft
func (p *Post) Unpublish() {
	p.Status = "draft"
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
	return p.Status == "published" && p.PublishedAt != nil
}

// IsDraft checks if the post is in draft status
func (p *Post) IsDraft() bool {
	return p.Status == "draft"
} 