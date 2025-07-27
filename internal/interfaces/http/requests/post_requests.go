package requests

import (
	"errors"
	"regexp"

	"github.com/google/uuid"
)

// CreatePostRequest represents the request for creating a post
type CreatePostRequest struct {
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Slug     string    `json:"slug"`
	Status   string    `json:"status"`
	AuthorID uuid.UUID `json:"author_id"`
}

// Validate validates the CreatePostRequest
func (r *CreatePostRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	if r.Slug == "" {
		return errors.New("slug is required")
	}
	if r.Status == "" {
		return errors.New("status is required")
	}
	if r.AuthorID == uuid.Nil {
		return errors.New("author_id is required")
	}

	// Validate slug format (alphanumeric and hyphens only)
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !slugRegex.MatchString(r.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate status
	if r.Status != "draft" && r.Status != "published" {
		return errors.New("status must be either 'draft' or 'published'")
	}

	return nil
}

// UpdatePostRequest represents the request for updating a post
type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Slug    string `json:"slug"`
	Status  string `json:"status"`
}

// Validate validates the UpdatePostRequest
func (r *UpdatePostRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	if r.Slug == "" {
		return errors.New("slug is required")
	}
	if r.Status == "" {
		return errors.New("status is required")
	}

	// Validate slug format (alphanumeric and hyphens only)
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !slugRegex.MatchString(r.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate status
	if r.Status != "draft" && r.Status != "published" {
		return errors.New("status must be either 'draft' or 'published'")
	}

	return nil
}
