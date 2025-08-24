package requests

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreatePostRequest represents the request for creating a post
type CreatePostRequest struct {
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Subtitle    string     `json:"subtitle"`
	Description string     `json:"description"`
	Language    string     `json:"language"`
	Layout      string     `json:"layout"`
	Content     string     `json:"content"`
	IsSticky    bool       `json:"is_sticky"`
	PublishedAt *time.Time `json:"published_at"`
}

// Validate validates the CreatePostRequest
func (r *CreatePostRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// ToEntity transforms CreatePostRequest to a Post entity
func (r *CreatePostRequest) ToEntity(authorID uuid.UUID) (*entities.Post, error) {
	// Generate slug from title if not provided
	slug := r.Slug
	if slug == "" {
		slug = generateSlug(r.Title)
	}

	// Create post entity
	post, err := entities.NewPost(
		r.Title,
		slug,
		r.Subtitle,
		r.Description,
		"post", // Default type
		r.Language,
		r.Layout,
		r.Content,
		authorID,
		r.IsSticky,
		r.PublishedAt,
	)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// UpdatePostRequest represents the request for updating a post
type UpdatePostRequest struct {
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Subtitle    string     `json:"subtitle"`
	Description string     `json:"description"`
	Language    string     `json:"language"`
	Layout      string     `json:"layout"`
	Content     string     `json:"content"`
	IsSticky    bool       `json:"is_sticky"`
	PublishedAt *time.Time `json:"published_at"`
}

// Validate validates the UpdatePostRequest
func (r *UpdatePostRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// ToEntity transforms UpdatePostRequest to a Post entity
func (r *UpdatePostRequest) ToEntity(existingPost *entities.Post, updatedBy uuid.UUID) (*entities.Post, error) {
	// Generate slug from title if not provided
	slug := r.Slug
	if slug == "" {
		slug = generateSlug(r.Title)
	}

	// Update the existing post
	if err := existingPost.UpdatePost(
		r.Title,
		slug,
		r.Subtitle,
		r.Description,
		"post", // Default type
		r.Language,
		r.Layout,
		r.IsSticky,
		r.PublishedAt,
	); err != nil {
		return nil, err
	}

	// Set the updated_by field
	existingPost.UpdatedBy = updatedBy

	return existingPost, nil
}

// generateSlug creates a URL-friendly slug from a title
func generateSlug(title string) string {
	// Simple slug generation - convert to lowercase and replace spaces with hyphens
	// In production, you might want more sophisticated slug generation
	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	// Remove special characters and keep only alphanumeric and hyphens
	// This is a basic implementation - you might want to use a proper slug library
	return slug
}
