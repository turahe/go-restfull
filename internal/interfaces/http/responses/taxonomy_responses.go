// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// TaxonomyResource represents a single taxonomy in API responses.
// This struct follows the Laravel API Resource pattern for consistent formatting
// and provides a comprehensive view of taxonomy data including hierarchical structure,
// nested set model information, and computed properties for tree operations.
type TaxonomyResource struct {
	// ID is the unique identifier for the taxonomy
	ID string `json:"id"`
	// Name is the display name of the taxonomy
	Name string `json:"name"`
	// Slug is the URL-friendly version of the taxonomy name
	Slug string `json:"slug"`
	// Code is an optional code identifier for the taxonomy
	Code string `json:"code,omitempty"`
	// Description is an optional description of the taxonomy's purpose
	Description string `json:"description,omitempty"`
	// ParentID is the optional ID of the parent taxonomy for hierarchical structures
	ParentID *string `json:"parent_id,omitempty"`
	// RecordLeft is used for nested set model operations (tree structure)
	RecordLeft *uint64 `json:"record_left,omitempty"`
	// RecordRight is used for nested set model operations (tree structure)
	RecordRight *uint64 `json:"record_right,omitempty"`
	// RecordDepth indicates the nesting level of the taxonomy in the tree
	RecordDepth *uint64 `json:"record_depth,omitempty"`
	// IsRoot indicates whether this taxonomy is at the root level (no parent)
	IsRoot bool `json:"is_root"`
	// HasChildren indicates whether this taxonomy has child taxonomies
	HasChildren bool `json:"has_children"`
	// HasParent indicates whether this taxonomy has a parent taxonomy
	HasParent bool `json:"has_parent"`
	// Level indicates the hierarchical level of the taxonomy (0 for root)
	Level int `json:"level"`
	// CreatedAt is the timestamp when the taxonomy was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the taxonomy was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the taxonomy was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	// Parent contains the parent taxonomy if this is a child taxonomy
	Parent *TaxonomyResource `json:"parent,omitempty"`
	// Children contains the nested child taxonomies
	Children []TaxonomyResource `json:"children,omitempty"`
}

// TaxonomyCollection represents a collection of taxonomies.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type TaxonomyCollection struct {
	// Data contains the array of taxonomy resources
	Data []TaxonomyResource `json:"data"`
	// Meta contains optional collection metadata (pagination, counts, etc.)
	Meta *CollectionMeta `json:"meta,omitempty"`
	// Links contains optional navigation links (first, last, prev, next)
	Links *CollectionLinks `json:"links,omitempty"`
}

// TaxonomyResourceResponse represents a single taxonomy response with Laravel-style formatting.
// This wrapper provides a consistent response structure with status information
// and follows the standard API response format used throughout the application.
type TaxonomyResourceResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the taxonomy resource
	Data TaxonomyResource `json:"data"`
}

// TaxonomyCollectionResponse represents a collection response with Laravel-style formatting.
// This wrapper provides a consistent response structure for collections with status information
// and follows the standard API response format used throughout the application.
type TaxonomyCollectionResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the taxonomy collection
	Data TaxonomyCollection `json:"data"`
}

// NewTaxonomyResource creates a new TaxonomyResource from a Taxonomy entity.
// This function transforms the domain entity into a consistent API response format,
// handling all optional fields, computed properties, and nested resources.
//
// Parameters:
//   - taxonomy: The taxonomy domain entity to convert
//
// Returns:
//   - A pointer to the newly created TaxonomyResource
func NewTaxonomyResource(taxonomy *entities.Taxonomy) *TaxonomyResource {
	// Handle optional parent ID for hierarchical taxonomies
	var parentID *string
	if taxonomy.ParentID != nil {
		parentIDStr := taxonomy.ParentID.String()
		parentID = &parentIDStr
	}

	// Transform parent taxonomy if exists
	var parentResource *TaxonomyResource
	if taxonomy.Parent != nil {
		parentResource = NewTaxonomyResource(taxonomy.Parent)
	}

	// Transform children taxonomies if exist
	var childrenResources []TaxonomyResource
	if taxonomy.Children != nil && len(taxonomy.Children) > 0 {
		childrenResources = make([]TaxonomyResource, len(taxonomy.Children))
		for i, child := range taxonomy.Children {
			childrenResources[i] = *NewTaxonomyResource(child)
		}
	}

	// Calculate computed level field from nested set model data
	level := 0
	if taxonomy.RecordDepth != nil {
		level = int(*taxonomy.RecordDepth)
	}

	return &TaxonomyResource{
		ID:          taxonomy.ID.String(),
		Name:        taxonomy.Name,
		Slug:        taxonomy.Slug,
		Code:        taxonomy.Code,
		Description: taxonomy.Description,
		ParentID:    parentID,
		RecordLeft:  taxonomy.RecordLeft,
		RecordRight: taxonomy.RecordRight,
		RecordDepth: taxonomy.RecordDepth,
		IsRoot:      taxonomy.ParentID == nil,
		HasChildren: taxonomy.Children != nil && len(taxonomy.Children) > 0,
		HasParent:   taxonomy.ParentID != nil,
		Level:       level,
		CreatedAt:   taxonomy.CreatedAt,
		UpdatedAt:   taxonomy.UpdatedAt,
		DeletedAt:   taxonomy.DeletedAt,
		Parent:      parentResource,
		Children:    childrenResources,
	}
}

// NewTaxonomyCollection creates a new TaxonomyCollection from a slice of Taxonomy entities.
// This function transforms multiple domain entities into a consistent API response format,
// creating a collection that can be easily serialized to JSON.
//
// Parameters:
//   - taxonomies: Slice of taxonomy domain entities to convert
//
// Returns:
//   - A pointer to the newly created TaxonomyCollection
func NewTaxonomyCollection(taxonomies []*entities.Taxonomy) *TaxonomyCollection {
	taxonomyResources := make([]TaxonomyResource, len(taxonomies))
	for i, taxonomy := range taxonomies {
		taxonomyResources[i] = *NewTaxonomyResource(taxonomy)
	}

	return &TaxonomyCollection{
		Data: taxonomyResources,
	}
}

// NewPaginatedTaxonomyCollection creates a new TaxonomyCollection with pagination metadata.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - taxonomies: Slice of taxonomy domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated TaxonomyCollection
func NewPaginatedTaxonomyCollection(
	taxonomies []*entities.Taxonomy,
	page, perPage int,
	total int64,
	baseURL string,
) *TaxonomyCollection {
	collection := NewTaxonomyCollection(taxonomies)

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

// NewTaxonomyResourceResponse creates a new single taxonomy response.
// This function wraps a TaxonomyResource in a standard API response format
// with a success status message.
//
// Parameters:
//   - taxonomy: The taxonomy domain entity to convert and wrap
//
// Returns:
//   - A pointer to the newly created TaxonomyResourceResponse
func NewTaxonomyResourceResponse(taxonomy *entities.Taxonomy) *TaxonomyResourceResponse {
	return &TaxonomyResourceResponse{
		Status: "success",
		Data:   *NewTaxonomyResource(taxonomy),
	}
}

// NewTaxonomyCollectionResponse creates a new taxonomy collection response.
// This function wraps a TaxonomyCollection in a standard API response format
// with a success status message.
//
// Parameters:
//   - taxonomies: Slice of taxonomy domain entities to convert and wrap
//
// Returns:
//   - A pointer to the newly created TaxonomyCollectionResponse
func NewTaxonomyCollectionResponse(taxonomies []*entities.Taxonomy) *TaxonomyCollectionResponse {
	return &TaxonomyCollectionResponse{
		Status: "success",
		Data:   *NewTaxonomyCollection(taxonomies),
	}
}

// NewPaginatedTaxonomyCollectionResponse creates a new paginated taxonomy collection response.
// This function wraps a paginated TaxonomyCollection in a standard API response format
// with a success status message and includes all pagination metadata.
//
// Parameters:
//   - taxonomies: Slice of taxonomy domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated TaxonomyCollectionResponse
func NewPaginatedTaxonomyCollectionResponse(
	taxonomies []*entities.Taxonomy,
	page, perPage int,
	total int64,
	baseURL string,
) *TaxonomyCollectionResponse {
	return &TaxonomyCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedTaxonomyCollection(taxonomies, page, perPage, total, baseURL),
	}
}
