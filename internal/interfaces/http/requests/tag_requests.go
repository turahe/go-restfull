package requests

import (
	"errors"
	"regexp"
)

// CreateTagRequest represents the request for creating a tag
type CreateTagRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// Validate validates the CreateTagRequest
func (r *CreateTagRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Slug == "" {
		return errors.New("slug is required")
	}

	// Validate slug format (alphanumeric and hyphens only)
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !slugRegex.MatchString(r.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate color format (hex color)
	if r.Color != "" {
		colorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
		if !colorRegex.MatchString(r.Color) {
			return errors.New("color must be a valid hex color (e.g., #FF0000)")
		}
	}

	return nil
}

// UpdateTagRequest represents the request for updating a tag
type UpdateTagRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// Validate validates the UpdateTagRequest
func (r *UpdateTagRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Slug == "" {
		return errors.New("slug is required")
	}

	// Validate slug format (alphanumeric and hyphens only)
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !slugRegex.MatchString(r.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate color format (hex color)
	if r.Color != "" {
		colorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
		if !colorRegex.MatchString(r.Color) {
			return errors.New("color must be a valid hex color (e.g., #FF0000)")
		}
	}

	return nil
}

// SearchTagsRequest represents the request for searching tags
type SearchTagsRequest struct {
	Query  string `json:"query" query:"query"`
	Limit  int    `json:"limit" query:"limit"`
	Offset int    `json:"offset" query:"offset"`
}

// Validate validates the SearchTagsRequest
func (r *SearchTagsRequest) Validate() error {
	if r.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if r.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	if r.Limit > 100 {
		return errors.New("limit cannot exceed 100")
	}
	return nil
}
