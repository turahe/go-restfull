// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/helper/pagination"

	"github.com/google/uuid"
)

// TaxonomyDTO represents a taxonomy in API responses.
// This struct provides a comprehensive view of taxonomy data including hierarchical
// organization using nested set model fields and optional parent-child relationships.
type TaxonomyDTO struct {
	// ID is the unique identifier for the taxonomy
	ID uuid.UUID `json:"id"`
	// Name is the display name of the taxonomy
	Name string `json:"name"`
	// Slug is the URL-friendly identifier for the taxonomy
	Slug string `json:"slug"`
	// Code is an optional short code for the taxonomy
	Code string `json:"code,omitempty"`
	// Description is an optional detailed description of the taxonomy
	Description string `json:"description,omitempty"`
	// ParentID is the optional ID of the parent taxonomy in the hierarchy
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	// RecordLeft is the left boundary value for nested set model hierarchy
	RecordLeft uint64 `json:"record_left"`
	// RecordRight is the right boundary value for nested set model hierarchy
	RecordRight uint64 `json:"record_right"`
	// RecordDepth is the depth level in the hierarchy tree
	RecordDepth uint64 `json:"record_depth"`
	// CreatedAt is the timestamp when the taxonomy was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the taxonomy was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the taxonomy was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Relationships
	// Parent contains the parent taxonomy if this taxonomy has a parent
	Parent *TaxonomyDTO `json:"parent,omitempty"`
	// Children contains the child taxonomies if this taxonomy has children
	Children []*TaxonomyDTO `json:"children,omitempty"`
}

// TaxonomySearchRequest represents the request for searching taxonomies.
// This struct defines the parameters for taxonomy search operations including
// query terms, pagination settings, and sorting preferences.
type TaxonomySearchRequest struct {
	// Query is the search term to filter taxonomies by name or description
	Query string `json:"query" query:"query"`
	// Page is the current page number for pagination (1-based)
	Page int `json:"page" query:"page"`
	// PerPage is the number of items to return per page
	PerPage int `json:"per_page" query:"per_page"`
	// SortBy specifies the field to sort results by
	SortBy string `json:"sort_by" query:"sort_by"`
	// SortDesc indicates whether to sort in descending order (true) or ascending (false)
	SortDesc bool `json:"sort_desc" query:"sort_desc"`
}

// TaxonomySearchResponse represents the response for taxonomy search with pagination.
// This struct provides search results along with pagination metadata for
// navigating through large result sets.
type TaxonomySearchResponse struct {
	// Data contains the array of taxonomy DTOs for the current page
	Data []*TaxonomyDTO `json:"data"`
	// Pagination contains metadata about the current page, total items, and navigation
	Pagination pagination.PaginationResponse `json:"pagination"`
}

// CreateTaxonomyRequest represents the request for creating a taxonomy.
// This struct defines the required and optional fields for taxonomy creation,
// with validation tags for required fields.
type CreateTaxonomyRequest struct {
	// Name is the display name of the taxonomy (required)
	Name string `json:"name" validate:"required"`
	// Slug is the URL-friendly identifier for the taxonomy (required)
	Slug string `json:"slug" validate:"required"`
	// Code is an optional short code for the taxonomy
	Code string `json:"code,omitempty"`
	// Description is an optional detailed description of the taxonomy
	Description string `json:"description,omitempty"`
	// ParentID is the optional ID of the parent taxonomy in the hierarchy
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// UpdateTaxonomyRequest represents the request for updating a taxonomy.
// This struct defines the fields that can be updated for an existing taxonomy,
// with validation tags for required fields.
type UpdateTaxonomyRequest struct {
	// Name is the display name of the taxonomy (required)
	Name string `json:"name" validate:"required"`
	// Slug is the URL-friendly identifier for the taxonomy (required)
	Slug string `json:"slug" validate:"required"`
	// Code is an optional short code for the taxonomy
	Code string `json:"code,omitempty"`
	// Description is an optional detailed description of the taxonomy
	Description string `json:"description,omitempty"`
	// ParentID is the optional ID of the parent taxonomy in the hierarchy
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// TaxonomyHierarchyResponse represents the response for taxonomy hierarchy.
// This struct provides hierarchical taxonomy data with pagination support
// for displaying tree structures in the UI.
type TaxonomyHierarchyResponse struct {
	// Data contains the array of taxonomy DTOs organized in hierarchy order
	Data []*TaxonomyDTO `json:"data"`
	// Pagination contains metadata about the current page, total items, and navigation
	Pagination pagination.PaginationResponse `json:"pagination"`
}

// TaxonomyDetailResponse represents the response for a single taxonomy with full details.
// This struct provides comprehensive information about a single taxonomy
// including all its fields and relationships.
type TaxonomyDetailResponse struct {
	// Data contains the complete taxonomy DTO with all details
	Data *TaxonomyDTO `json:"data"`
}
