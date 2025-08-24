package requests

import (
	"errors"
	"time"
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

	return nil
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

	return nil
}
