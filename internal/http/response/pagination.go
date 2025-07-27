package response

import (
	"math"
)

// PaginationResponse represents the pagination metadata in API responses
type PaginationResponse struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	HasNextPage  bool  `json:"has_next_page"`
	HasPrevPage  bool  `json:"has_prev_page"`
	NextPage     int   `json:"next_page,omitempty"`
	PreviousPage int   `json:"previous_page,omitempty"`
	From         int   `json:"from"`
	To           int   `json:"to"`
}

// PaginatedResult represents a paginated response with data and metadata
type PaginatedResult struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// CreatePaginationResponse creates a pagination response from request parameters and total count
func CreatePaginationResponse(page, perPage int, totalItems int64) PaginationResponse {
	totalPages := int(math.Ceil(float64(totalItems) / float64(perPage)))

	// Ensure current page doesn't exceed total pages
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	from := (page-1)*perPage + 1
	to := page * perPage
	if to > int(totalItems) {
		to = int(totalItems)
	}
	if totalItems == 0 {
		from = 0
		to = 0
	}

	hasNextPage := page < totalPages
	hasPrevPage := page > 1

	var nextPage, prevPage int
	if hasNextPage {
		nextPage = page + 1
	}
	if hasPrevPage {
		prevPage = page - 1
	}

	return PaginationResponse{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		HasNextPage:  hasNextPage,
		HasPrevPage:  hasPrevPage,
		NextPage:     nextPage,
		PreviousPage: prevPage,
		From:         from,
		To:           to,
	}
}

// CreatePaginatedResult creates a complete paginated result
func CreatePaginatedResult(data interface{}, page, perPage int, totalItems int64) *PaginatedResult {
	return &PaginatedResult{
		Data:       data,
		Pagination: CreatePaginationResponse(page, perPage, totalItems),
	}
}

// Legacy pagination response (keeping for backward compatibility)
type PaginationResponseLegacy struct {
	Data         interface{} `json:"data"`
	TotalCount   int         `json:"total_count"`
	TotalPage    int         `json:"total_page"`
	CurrentPage  int         `json:"current_page"`
	LastPage     int         `json:"last_page"`
	PerPage      int         `json:"per_page"`
	NextPage     int         `json:"next_page"`
	PreviousPage int         `json:"previous_page"`
	Path         string      `json:"path"`
}
