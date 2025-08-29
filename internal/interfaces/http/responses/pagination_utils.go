// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

// PaginationHelper provides utility functions for creating paginated collections
// and responses to reduce code duplication across response files.
type PaginationHelper struct{}

// NewPaginationHelper creates a new pagination helper instance
func NewPaginationHelper() *PaginationHelper {
	return &PaginationHelper{}
}

// CreatePaginatedCollection creates a paginated collection with metadata and links.
// This function consolidates the common pagination logic used across all response files.
//
// Parameters:
//   - data: The data array for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A CollectionMeta with pagination metadata
//   - A CollectionLinks with navigation links
func (h *PaginationHelper) CreatePaginatedCollection(page, perPage, total int, baseURL string) (CollectionMeta, CollectionLinks) {
	// Calculate pagination metadata
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

	// Create metadata
	meta := CollectionMeta{
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

	// Create navigation links
	links := CollectionLinks{
		First: buildPaginationLink(baseURL, 1, perPage),
		Last:  buildPaginationLink(baseURL, totalPages, perPage),
	}

	// Add previous page link if not on first page
	if page > 1 {
		links.Prev = buildPaginationLink(baseURL, page-1, perPage)
	}

	// Add next page link if not on last page
	if page < totalPages {
		links.Next = buildPaginationLink(baseURL, page+1, perPage)
	}

	return meta, links
}

// CreatePaginatedCollectionResponse creates a paginated collection response.
// This function provides a generic way to create paginated responses for any collection type.
//
// Parameters:
//   - data: The collection data
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//   - message: Success message for the response
//
// Returns:
//   - A GenericResponse with paginated data and metadata
func (h *PaginationHelper) CreatePaginatedCollectionResponse(data interface{}, page, perPage, total int, baseURL string, message string) GenericResponse {
	meta, links := h.CreatePaginatedCollection(page, perPage, total, baseURL)

	collection := PaginatedResult{
		Data:  data,
		Meta:  meta,
		Links: links,
	}

	return NewGenericResponse(message, collection)
}

// Global pagination helper instance for easy access
var DefaultPaginationHelper = NewPaginationHelper()

// CreatePaginatedCollection is a convenience function that uses the default pagination helper
func CreatePaginatedCollection(page, perPage, total int, baseURL string) (CollectionMeta, CollectionLinks) {
	return DefaultPaginationHelper.CreatePaginatedCollection(page, perPage, total, baseURL)
}

// CreatePaginatedCollectionResponse is a convenience function that uses the default pagination helper
func CreatePaginatedCollectionResponse(data interface{}, page, perPage, total int, baseURL string, message string) GenericResponse {
	return DefaultPaginationHelper.CreatePaginatedCollectionResponse(data, page, perPage, total, baseURL, message)
}
