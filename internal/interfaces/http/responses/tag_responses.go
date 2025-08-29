// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// TagResource represents a tag in API responses.
// This struct provides a comprehensive view of tag data including basic information,
// visual properties, audit trail, and computed properties for easy status checking.
// It follows the Laravel API Resource pattern for consistent formatting.
type TagResource struct {
	// ID is the unique identifier for the tag
	ID string `json:"id"`
	// Name is the display name of the tag
	Name string `json:"name"`
	// Slug is the URL-friendly version of the tag name
	Slug string `json:"slug"`
	// Description provides additional context about the tag's purpose
	Description string `json:"description"`
	// Color is the visual color identifier for the tag (e.g., hex code, color name)
	Color string `json:"color"`
	// CreatedBy is the ID of the user who created the tag
	CreatedBy string `json:"created_by"`
	// UpdatedBy is the ID of the user who last updated the tag
	UpdatedBy string `json:"updated_by"`
	// DeletedBy is the optional ID of the user who deleted the tag
	DeletedBy *string `json:"deleted_by,omitempty"`
	// CreatedAt is the timestamp when the tag was created
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the tag was last updated
	UpdatedAt string `json:"updated_at"`
	// DeletedAt is the optional timestamp when the tag was soft-deleted
	DeletedAt *string `json:"deleted_at,omitempty"`

	// Computed fields for easy status checking
	// IsDeleted indicates whether the tag has been soft-deleted
	IsDeleted bool `json:"is_deleted"`
}

// TagCollection represents a collection of tags.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type TagCollection struct {
	// Data contains the array of tag resources
	Data []TagResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// TagResourceResponse represents a single tag response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type TagResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the tag resource
	Data TagResource `json:"data"`
}

// TagCollectionResponse represents a collection of tags response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type TagCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the tag collection
	Data TagCollection `json:"data"`
}

// NewTagResource creates a new TagResource from tag entity.
// This function transforms the domain entity into a consistent API response format,
// handling all optional fields and computed properties appropriately.
//
// Parameters:
//   - tag: The tag domain entity to convert
//
// Returns:
//   - A new TagResource with all fields properly formatted
func NewTagResource(tag *entities.Tag) TagResource {
	resource := TagResource{
		ID:          tag.ID.String(),
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		Color:       tag.Color,
		CreatedBy:   tag.CreatedBy.String(),
		UpdatedBy:   tag.UpdatedBy.String(),
		CreatedAt:   tag.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   tag.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),

		// Set computed fields based on tag state
		IsDeleted: tag.IsDeleted(),
	}

	// Handle optional soft deletion information
	if tag.DeletedBy != nil {
		deletedBy := tag.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if tag.DeletedAt != nil {
		deletedAt := tag.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	return resource
}

// NewTagResourceResponse creates a new TagResourceResponse.
// This function wraps a TagResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - tag: The tag domain entity to convert and wrap
//
// Returns:
//   - A new TagResourceResponse with success status and tag data
func NewTagResourceResponse(tag *entities.Tag) TagResourceResponse {
	return TagResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Tag retrieved successfully",
		Data:            NewTagResource(tag),
	}
}

// NewTagCollection creates a new TagCollection.
// This function transforms multiple tag domain entities into a consistent
// API response format, creating a collection that can be easily serialized to JSON.
//
// Parameters:
//   - tags: Slice of tag domain entities to convert
//
// Returns:
//   - A new TagCollection with all tags properly formatted
func NewTagCollection(tags []*entities.Tag) TagCollection {
	tagResources := make([]TagResource, len(tags))
	for i, tag := range tags {
		tagResources[i] = NewTagResource(tag)
	}

	return TagCollection{
		Data: tagResources,
	}
}

// NewPaginatedTagCollection creates a new TagCollection with pagination.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - tags: Slice of tag domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated TagCollection with metadata and navigation links
func NewPaginatedTagCollection(tags []*entities.Tag, page, perPage, total int, baseURL string) TagCollection {
	collection := NewTagCollection(tags)

	// Calculate pagination metadata
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

	// Set pagination metadata
	collection.Meta = CollectionMeta{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   int64(total),
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		From:         from,
		To:           to,
	}

	// Build pagination navigation links
	collection.Links = CollectionLinks{
		First: buildPaginationLink(baseURL, 1, perPage),
		Last:  buildPaginationLink(baseURL, totalPages, perPage),
	}

	// Add previous page link if not on first page
	if page > 1 {
		collection.Links.Prev = buildPaginationLink(baseURL, page-1, perPage)
	}

	// Add next page link if not on last page
	if page < totalPages {
		collection.Links.Next = buildPaginationLink(baseURL, page+1, perPage)
	}

	return collection
}

// NewTagCollectionResponse creates a new TagCollectionResponse.
// This function wraps a TagCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - tags: Slice of tag domain entities to convert and wrap
//
// Returns:
//   - A new TagCollectionResponse with success status and tag collection data
func NewTagCollectionResponse(tags []*entities.Tag) TagCollectionResponse {
	return TagCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Tags retrieved successfully",
		Data:            NewTagCollection(tags),
	}
}

// NewPaginatedTagCollectionResponse creates a new TagCollectionResponse with pagination.
// This function wraps a paginated TagCollection in a standard API response format
// with appropriate response codes and success messages, including all pagination metadata.
//
// Parameters:
//   - tags: Slice of tag domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated TagCollectionResponse with success status and pagination data
func NewPaginatedTagCollectionResponse(tags []*entities.Tag, page, perPage, total int, baseURL string) TagCollectionResponse {
	return TagCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Tags retrieved successfully",
		Data:            NewPaginatedTagCollection(tags, page, perPage, total, baseURL),
	}
}
