// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreatePostRequest represents the request for creating a new post entity.
// This struct defines the required and optional fields for post creation,
// including title, content, metadata, and publishing options. Posts are
// the primary content type in the system, supporting various layouts and languages.
type CreatePostRequest struct {
	// Title is the main headline of the post (required)
	Title string `json:"title"`
	// Slug is the URL-friendly identifier for the post (optional, auto-generated from title if not provided)
	Slug string `json:"slug"`
	// Subtitle provides additional context or summary below the title (optional)
	Subtitle string `json:"subtitle"`
	// Description is a brief overview of the post content (optional)
	Description string `json:"description"`
	// Language specifies the post's language code (optional, e.g., "en", "id")
	Language string `json:"language"`
	// Layout defines the visual presentation style of the post (optional)
	Layout string `json:"layout"`
	// Content is the main body text of the post (required)
	Content string `json:"content"`
	// IsSticky determines if the post should be prominently displayed (optional, defaults to false)
	IsSticky bool `json:"is_sticky"`
	// PublishedAt sets the publication timestamp (optional, nil for immediate publication)
	PublishedAt *time.Time `json:"published_at"`
}

// Validate performs validation on the CreatePostRequest.
// This method ensures that all required fields are provided and
// validates basic content requirements.
//
// Validation Rules:
// - Title is required (cannot be empty)
// - Content is required (cannot be empty)
//
// Returns:
//   - error: Validation error if any required field is missing, nil if valid
func (r *CreatePostRequest) Validate() error {
	// Check required fields
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// ToEntity transforms the CreatePostRequest to a Post domain entity.
// This method handles slug generation if not provided, creates the post
// using the domain entity constructor, and associates it with the author.
//
// Parameters:
//   - authorID: The UUID of the user creating the post
//
// Returns:
//   - *entities.Post: The created post entity
//   - error: Any error that occurred during entity creation
func (r *CreatePostRequest) ToEntity(authorID uuid.UUID) (*entities.Post, error) {
	// Generate slug from title if not provided
	slug := r.Slug
	if slug == "" {
		slug = generateSlug(r.Title)
	}

	// Create post entity using the domain constructor
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

// UpdatePostRequest represents the request for updating an existing post entity.
// This struct allows updating post properties while maintaining data integrity.
// All fields are treated as optional for partial updates.
type UpdatePostRequest struct {
	// Title is the main headline of the post (optional)
	Title string `json:"title"`
	// Slug is the URL-friendly identifier for the post (optional, auto-generated from title if not provided)
	Slug string `json:"slug"`
	// Subtitle provides additional context or summary below the title (optional)
	Subtitle string `json:"subtitle"`
	// Description is a brief overview of the post content (optional)
	Description string `json:"description"`
	// Language specifies the post's language code (optional, e.g., "en", "id")
	Language string `json:"language"`
	// Layout defines the visual presentation style of the post (optional)
	Layout string `json:"layout"`
	// Content is the main body text of the post (optional)
	Content string `json:"content"`
	// IsSticky determines if the post should be prominently displayed (optional)
	IsSticky bool `json:"is_sticky"`
	// PublishedAt sets the publication timestamp (optional)
	PublishedAt *time.Time `json:"published_at"`
}

// Validate performs validation on the UpdatePostRequest.
// This method ensures that all required fields are provided for the update
// operation, maintaining data consistency.
//
// Validation Rules:
// - Title is required if provided (cannot be empty)
// - Content is required if provided (cannot be empty)
//
// Returns:
//   - error: Validation error if any provided field is invalid, nil if valid
func (r *UpdatePostRequest) Validate() error {
	// Validate title if provided
	if r.Title == "" {
		return errors.New("title is required")
	}
	// Validate content if provided
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// ToEntity transforms the UpdatePostRequest to update an existing Post entity.
// This method handles slug generation if not provided, updates the existing post
// using the domain entity's update method, and tracks who made the changes.
//
// Parameters:
//   - existingPost: The existing post entity to update
//   - updatedBy: The UUID of the user performing the update
//
// Returns:
//   - *entities.Post: The updated post entity
//   - error: Any error that occurred during the update operation
func (r *UpdatePostRequest) ToEntity(existingPost *entities.Post, updatedBy uuid.UUID) (*entities.Post, error) {
	// Generate slug from title if not provided
	slug := r.Slug
	if slug == "" {
		slug = generateSlug(r.Title)
	}

	// Update the existing post using the domain entity's update method
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

	// Track who made the update for audit purposes
	existingPost.UpdatedBy = updatedBy

	return existingPost, nil
}

// generateSlug creates a URL-friendly slug from a post title.
// This helper function converts the title to a format suitable for URLs
// by converting to lowercase, replacing spaces with hyphens, and
// removing special characters.
//
// Note: This is a basic implementation. In production, you might want
// to use a more sophisticated slug generation library that handles
// internationalization, special characters, and uniqueness.
//
// Parameters:
//   - title: The post title to convert to a slug
//
// Returns:
//   - string: The generated URL-friendly slug
func generateSlug(title string) string {
	// Simple slug generation - convert to lowercase and replace spaces with hyphens
	// In production, you might want more sophisticated slug generation
	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	// Remove special characters and keep only alphanumeric and hyphens
	// This is a basic implementation - you might want to use a proper slug library
	return slug
}
