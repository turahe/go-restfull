// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"fmt"
	"net/url"
	"strconv"
)

// Common response codes for consistent API responses
const (
	SYSTEM_OPERATION_SUCCESS = 200
	BAD_REQUEST              = 400
	UNAUTHORIZED             = 401
	FORBIDDEN                = 403
	NOT_FOUND                = 404
	INTERNAL_SERVER_ERROR    = 500
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

// CommonResponse represents a generic API response
type CommonResponse struct {
	ResponseCode    int         `json:"response_code"`
	ResponseMessage string      `json:"response_message"`
	Data            interface{} `json:"data,omitempty"`
	Errors          interface{} `json:"errors,omitempty"`
	RequestID       string      `json:"request_id,omitempty"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	HasNextPage  bool  `json:"has_next_page"`
	HasPrevPage  bool  `json:"has_prev_page"`
	NextPage     int   `json:"next_page"`
	PreviousPage int   `json:"previous_page"`
	From         int   `json:"from"`
	To           int   `json:"to"`
}

// PaginatedResult represents a paginated result with metadata
type PaginatedResult struct {
	Data  interface{}     `json:"data"`
	Meta  CollectionMeta  `json:"meta"`
	Links CollectionLinks `json:"links"`
}

// CollectionMeta represents collection metadata for pagination
type CollectionMeta struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	HasNextPage  bool  `json:"has_next_page"`
	HasPrevPage  bool  `json:"has_prev_page"`
	NextPage     int   `json:"next_page"`
	PreviousPage int   `json:"previous_page"`
	From         int   `json:"from"`
	To           int   `json:"to"`
}

// CollectionLinks represents navigation links for pagination
type CollectionLinks struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
}

// CreatePaginatedResult creates a paginated result with metadata and links
func CreatePaginatedResult(data interface{}, page, perPage, total int, baseURL string) PaginatedResult {
	collection := PaginatedResult{
		Data: data,
	}

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

// CreatePaginationResponse creates a pagination response with metadata
func CreatePaginationResponse(page, perPage, total int) PaginationResponse {
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

	return PaginationResponse{
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
}

// CreateCollectionMeta creates collection metadata for pagination
func CreateCollectionMeta(page, perPage, total int) CollectionMeta {
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

	return CollectionMeta{
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
}

// CreateCollectionLinks creates navigation links for pagination
func CreateCollectionLinks(baseURL string, page, perPage, total int) CollectionLinks {
	totalPages := (total + perPage - 1) / perPage

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

	return links
}

// buildPaginationLink builds a pagination link with query parameters.
// This function constructs URLs with proper query parameters for pagination,
// preserving existing query parameters while updating page-specific ones.
func buildPaginationLink(baseURL string, page, perPage int) string {
	// Parse the base URL to extract existing query parameters
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// If parsing fails, return a simple URL with page and per_page parameters
		return fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page, perPage)
	}

	// Get existing query parameters
	query := parsedURL.Query()

	// Update or add pagination parameters
	query.Set("page", strconv.Itoa(page))
	query.Set("per_page", strconv.Itoa(perPage))

	// Reconstruct the URL with updated query parameters
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}

// generatePageURL generates a pagination URL for a specific page.
// This function creates a URL with only the page parameter, useful for
// endpoints that use a default limit. It returns an empty string for invalid page numbers.
//
// Parameters:
//   - baseURL: The base URL for the endpoint
//   - page: The page number to link to
//
// Returns:
//   - A formatted URL with the page query parameter, or empty string if invalid
func generatePageURL(baseURL string, page int) string {
	if page <= 0 {
		return ""
	}
	return baseURL + "?page=" + strconv.Itoa(page)
}

// GenericResponse represents a generic API response wrapper
type GenericResponse struct {
	ResponseCode    int         `json:"response_code"`
	ResponseMessage string      `json:"response_message"`
	Data            interface{} `json:"data"`
}

// NewGenericResponse creates a new generic response with success status
func NewGenericResponse(message string, data interface{}) GenericResponse {
	return GenericResponse{
		ResponseCode:    SYSTEM_OPERATION_SUCCESS,
		ResponseMessage: message,
		Data:            data,
	}
}

// NewGenericErrorResponse creates a new generic error response
func NewGenericErrorResponse(code int, message string, errors interface{}) GenericResponse {
	return GenericResponse{
		ResponseCode:    code,
		ResponseMessage: message,
		Data:            errors,
	}
}
