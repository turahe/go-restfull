// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// PostResource represents a single post in API responses.
// This struct follows the Laravel API Resource pattern for consistent formatting
// and provides a comprehensive view of post data including content, metadata,
// and publication status.
type PostResource struct {
	// ID is the unique identifier for the post
	ID string `json:"id"`
	// Title is the main title of the post
	Title string `json:"title"`
	// Slug is the URL-friendly version of the post title
	Slug string `json:"slug"`
	// Subtitle is an optional subtitle for the post
	Subtitle string `json:"subtitle"`
	// Description is a brief description or excerpt of the post content
	Description string `json:"description"`
	// Type indicates the type of post (e.g., "article", "news", "blog")
	Type string `json:"type"`
	// IsSticky indicates whether the post should be displayed prominently
	IsSticky bool `json:"is_sticky"`
	// Language specifies the language of the post content
	Language string `json:"language"`
	// Layout specifies the display layout for the post
	Layout string `json:"layout"`
	// Content contains the main body content of the post
	Content string `json:"content"`
	// PublishedAt is the optional timestamp when the post was published
	PublishedAt *time.Time `json:"published_at,omitempty"`
	// IsPublished indicates whether the post is currently published
	IsPublished bool `json:"is_published"`
	// Status indicates the current status of the post (e.g., "draft", "published")
	Status string `json:"status"`
	// CreatedAt is the timestamp when the post was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the post was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the post was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// PostCollection represents a collection of posts.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type PostCollection struct {
	// Data contains the array of post resources
	Data []PostResource `json:"data"`
	// Meta contains optional collection metadata (pagination, counts, etc.)
	Meta *CollectionMeta `json:"meta,omitempty"`
	// Links contains optional navigation links (first, last, prev, next)
	Links *CollectionLinks `json:"links,omitempty"`
}

// PostResourceResponse represents a single post response with Laravel-style formatting.
// This wrapper provides a consistent response structure with status information
// and follows the standard API response format used throughout the application.
type PostResourceResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the post resource
	Data PostResource `json:"data"`
}

// PostCollectionResponse represents a collection response with Laravel-style formatting.
// This wrapper provides a consistent response structure for collections with status information
// and follows the standard API response format used throughout the application.
type PostCollectionResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the post collection
	Data PostCollection `json:"data"`
}

// NewPostResource creates a new PostResource from a Post entity.
// This function transforms the domain entity into a consistent API response format,
// automatically determining the publication status based on the PublishedAt field.
//
// Parameters:
//   - post: The post domain entity to convert
//
// Returns:
//   - A pointer to the newly created PostResource
func NewPostResource(post *entities.Post) *PostResource {
	// Determine status based on published_at field
	status := "draft"
	if post.PublishedAt != nil {
		status = "published"
	}

	return &PostResource{
		ID:          post.ID.String(),
		Title:       post.Title,
		Slug:        post.Slug,
		Subtitle:    post.Subtitle,
		Description: post.Description,
		Type:        post.Type,
		IsSticky:    post.IsSticky,
		Language:    post.Language,
		Layout:      post.Layout,
		Content:     post.Content,
		PublishedAt: post.PublishedAt,
		IsPublished: post.PublishedAt != nil,
		Status:      status,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		DeletedAt:   post.DeletedAt,
	}
}

// NewPostCollection creates a new PostCollection from a slice of Post entities.
// This function transforms multiple domain entities into a consistent API response format,
// creating a collection that can be easily serialized to JSON.
//
// Parameters:
//   - posts: Slice of post domain entities to convert
//
// Returns:
//   - A pointer to the newly created PostCollection
func NewPostCollection(posts []*entities.Post) *PostCollection {
	postResources := make([]PostResource, len(posts))
	for i, post := range posts {
		postResources[i] = *NewPostResource(post)
	}

	return &PostCollection{
		Data: postResources,
	}
}

// NewPaginatedPostCollection creates a new PostCollection with pagination metadata.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - posts: Slice of post domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated PostCollection
func NewPaginatedPostCollection(
	posts []*entities.Post,
	page, perPage int,
	total int64,
	baseURL string,
) *PostCollection {
	collection := NewPostCollection(posts)

	// Calculate total pages with proper handling of edge cases
	totalPages := (int(total) + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Calculate the range of items on the current page
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > int(total) {
		to = int(total)
	}

	// Set pagination metadata
	collection.Meta = &CollectionMeta{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   total,
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		From:         from,
		To:           to,
	}

	// Generate pagination navigation links
	collection.Links = &CollectionLinks{
		First: generatePageURL(baseURL, 1),
		Last:  generatePageURL(baseURL, totalPages),
		Prev:  generatePageURL(baseURL, page-1),
		Next:  generatePageURL(baseURL, page+1),
	}

	return collection
}

// NewPostResourceResponse creates a new single post response.
// This function wraps a PostResource in a standard API response format
// with a success status message.
//
// Parameters:
//   - post: The post domain entity to convert and wrap
//
// Returns:
//   - A pointer to the newly created PostResourceResponse
func NewPostResourceResponse(post *entities.Post) *PostResourceResponse {
	return &PostResourceResponse{
		Status: "success",
		Data:   *NewPostResource(post),
	}
}

// NewPostCollectionResponse creates a new post collection response.
// This function wraps a PostCollection in a standard API response format
// with a success status message.
//
// Parameters:
//   - posts: Slice of post domain entities to convert and wrap
//
// Returns:
//   - A pointer to the newly created PostCollectionResponse
func NewPostCollectionResponse(posts []*entities.Post) *PostCollectionResponse {
	return &PostCollectionResponse{
		Status: "success",
		Data:   *NewPostCollection(posts),
	}
}

// NewPaginatedPostCollectionResponse creates a new paginated post collection response.
// This function wraps a paginated PostCollection in a standard API response format
// with a success status message and includes all pagination metadata.
//
// Parameters:
//   - posts: Slice of post domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated PostCollectionResponse
func NewPaginatedPostCollectionResponse(
	posts []*entities.Post,
	page, perPage int,
	total int64,
	baseURL string,
) *PostCollectionResponse {
	return &PostCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedPostCollection(posts, page, perPage, total, baseURL),
	}
}
