package response

import (
	"time"

	"github.com/turahe/go-restfull/internal/helper/pagination"

	"github.com/google/uuid"
)

// TaxonomyDTO represents a taxonomy in API responses
type TaxonomyDTO struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Code        string     `json:"code,omitempty"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	RecordLeft  int64      `json:"record_left"`
	RecordRight int64      `json:"record_right"`
	RecordDepth int64      `json:"record_depth"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`

	// Relationships
	Parent   *TaxonomyDTO   `json:"parent,omitempty"`
	Children []*TaxonomyDTO `json:"children,omitempty"`
}

// TaxonomySearchRequest represents the request for searching taxonomies
type TaxonomySearchRequest struct {
	Query    string `json:"query" query:"query"`
	Page     int    `json:"page" query:"page"`
	PerPage  int    `json:"per_page" query:"per_page"`
	SortBy   string `json:"sort_by" query:"sort_by"`
	SortDesc bool   `json:"sort_desc" query:"sort_desc"`
}

// TaxonomySearchResponse represents the response for taxonomy search with pagination
type TaxonomySearchResponse struct {
	Data       []*TaxonomyDTO                `json:"data"`
	Pagination pagination.PaginationResponse `json:"pagination"`
}

// CreateTaxonomyRequest represents the request for creating a taxonomy
type CreateTaxonomyRequest struct {
	Name        string     `json:"name" validate:"required"`
	Slug        string     `json:"slug" validate:"required"`
	Code        string     `json:"code,omitempty"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// UpdateTaxonomyRequest represents the request for updating a taxonomy
type UpdateTaxonomyRequest struct {
	Name        string     `json:"name" validate:"required"`
	Slug        string     `json:"slug" validate:"required"`
	Code        string     `json:"code,omitempty"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// TaxonomyHierarchyResponse represents the response for taxonomy hierarchy
type TaxonomyHierarchyResponse struct {
	Data       []*TaxonomyDTO                `json:"data"`
	Pagination pagination.PaginationResponse `json:"pagination"`
}

// TaxonomyDetailResponse represents the response for a single taxonomy with full details
type TaxonomyDetailResponse struct {
	Data *TaxonomyDTO `json:"data"`
}
