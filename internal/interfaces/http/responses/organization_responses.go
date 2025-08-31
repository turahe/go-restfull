// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// OrganizationResource represents a single organization in API responses.
// This struct follows the Laravel API Resource pattern for consistent formatting
// and provides a comprehensive view of organization data including hierarchical structure,
// nested set model information, and computed properties.
type OrganizationResource struct {
	// ID is the unique identifier for the organization
	ID string `json:"id"`
	// Name is the display name of the organization
	Name string `json:"name"`
	// Description is an optional description of the organization
	Description *string `json:"description,omitempty"`
	// Code is an optional code identifier for the organization
	Code *string `json:"code,omitempty"`
	// Type is an optional type classification for the organization
	Type *string `json:"type,omitempty"`
	// Status indicates the current status of the organization
	Status string `json:"status"`
	// ParentID is the optional ID of the parent organization for hierarchical structures
	ParentID *string `json:"parent_id,omitempty"`
	// RecordLeft is used for nested set model operations (tree structure)
	RecordLeft *int64 `json:"record_left,omitempty"`
	// RecordRight is used for nested set model operations (tree structure)
	RecordRight *int64 `json:"record_right,omitempty"`
	// RecordDepth indicates the nesting level of the organization in the tree
	RecordDepth *int64 `json:"record_depth,omitempty"`
	// RecordOrdering determines the display order of organizations
	RecordOrdering *int64 `json:"record_ordering,omitempty"`
	// IsRoot indicates whether this organization is at the root level (no parent)
	IsRoot bool `json:"is_root"`
	// HasChildren indicates whether this organization has child organizations
	HasChildren bool `json:"has_children"`
	// HasParent indicates whether this organization has a parent organization
	HasParent bool `json:"has_parent"`
	// Level indicates the hierarchical level of the organization (0 for root)
	Level int `json:"level"`
	// CreatedAt is the timestamp when the organization was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the organization was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the organization was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	// Parent contains the parent organization if this is a child organization
	Parent *OrganizationResource `json:"parent,omitempty"`
	// Children contains the nested child organizations
	Children []OrganizationResource `json:"children,omitempty"`
}

// OrganizationCollection represents a collection of organizations.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type OrganizationCollection struct {
	// Data contains the array of organization resources
	Data []OrganizationResource `json:"data"`
	// Meta contains optional collection metadata (pagination, counts, etc.)
	Meta *CollectionMeta `json:"meta,omitempty"`
	// Links contains optional navigation links (first, last, prev, next)
	Links *CollectionLinks `json:"links,omitempty"`
}

// OrganizationResourceResponse represents a single organization response with Laravel-style formatting.
// This wrapper provides a consistent response structure with status information
// and follows the standard API response format used throughout the application.
type OrganizationResourceResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the organization resource
	Data OrganizationResource `json:"data"`
}

// OrganizationCollectionResponse represents a collection response with Laravel-style formatting.
// This wrapper provides a consistent response structure for collections with status information
// and follows the standard API response format used throughout the application.
type OrganizationCollectionResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the organization collection
	Data OrganizationCollection `json:"data"`
}

// NewOrganizationResource creates a new OrganizationResource from an Organization entity.
// This function transforms the domain entity into a consistent API response format,
// handling all optional fields, computed properties, and nested resources.
//
// Parameters:
//   - org: The organization domain entity to convert
//
// Returns:
//   - A pointer to the newly created OrganizationResource
func NewOrganizationResource(org *entities.Organization) *OrganizationResource {
	// Handle optional parent ID for hierarchical organizations
	var parentID *string
	if org.ParentID != nil {
		parentIDStr := org.ParentID.String()
		parentID = &parentIDStr
	}

	// Handle optional organization type
	var orgType *string
	if org.Type != nil {
		typeStr := string(*org.Type)
		orgType = &typeStr
	}

	// Transform parent organization if exists
	var parentResource *OrganizationResource
	if org.Parent != nil {
		parentResource = NewOrganizationResource(org.Parent)
	}

	// Transform children organizations if exist
	var childrenResources []OrganizationResource
	if org.Children != nil && len(org.Children) > 0 {
		childrenResources = make([]OrganizationResource, len(org.Children))
		for i, child := range org.Children {
			childrenResources[i] = *NewOrganizationResource(child)
		}
	}

	// Calculate computed level field from nested set model data
	level := 0
	if org.RecordDepth != nil {
		level = int(*org.RecordDepth)
	}

	return &OrganizationResource{
		ID:             org.ID.String(),
		Name:           org.Name,
		Description:    org.Description,
		Code:           org.Code,
		Type:           orgType,
		Status:         string(org.Status),
		ParentID:       parentID,
		RecordLeft:     org.RecordLeft,
		RecordRight:    org.RecordRight,
		RecordDepth:    org.RecordDepth,
		RecordOrdering: org.RecordOrdering,
		IsRoot:         org.ParentID == nil,
		HasChildren:    org.Children != nil && len(org.Children) > 0,
		HasParent:      org.ParentID != nil,
		Level:          level,
		CreatedAt:      org.CreatedAt,
		UpdatedAt:      org.UpdatedAt,
		DeletedAt:      org.DeletedAt,
		Parent:         parentResource,
		Children:       childrenResources,
	}
}

// NewOrganizationCollection creates a new OrganizationCollection from a slice of Organization entities.
// This function transforms multiple domain entities into a consistent API response format,
// creating a collection that can be easily serialized to JSON.
//
// Parameters:
//   - organizations: Slice of organization domain entities to convert
//
// Returns:
//   - A pointer to the newly created OrganizationCollection
func NewOrganizationCollection(organizations []*entities.Organization) *OrganizationCollection {
	orgResources := make([]OrganizationResource, len(organizations))
	for i, org := range organizations {
		orgResources[i] = *NewOrganizationResource(org)
	}

	return &OrganizationCollection{
		Data: orgResources,
	}
}

// NewPaginatedOrganizationCollection creates a new OrganizationCollection with pagination metadata.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - organizations: Slice of organization domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated OrganizationCollection
func NewPaginatedOrganizationCollection(
	organizations []*entities.Organization,
	page, perPage int,
	total int64,
	baseURL string,
) *OrganizationCollection {
	collection := NewOrganizationCollection(organizations)

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

// NewOrganizationResourceResponse creates a new single organization response.
// This function wraps an OrganizationResource in a standard API response format
// with a success status message.
//
// Parameters:
//   - org: The organization domain entity to convert and wrap
//
// Returns:
//   - A pointer to the newly created OrganizationResourceResponse
func NewOrganizationResourceResponse(org *entities.Organization) *OrganizationResourceResponse {
	return &OrganizationResourceResponse{
		Status: "success",
		Data:   *NewOrganizationResource(org),
	}
}

// NewOrganizationCollectionResponse creates a new organization collection response.
// This function wraps an OrganizationCollection in a standard API response format
// with a success status message.
//
// Parameters:
//   - organizations: Slice of organization domain entities to convert and wrap
//
// Returns:
//   - A pointer to the newly created OrganizationCollectionResponse
func NewOrganizationCollectionResponse(organizations []*entities.Organization) *OrganizationCollectionResponse {
	return &OrganizationCollectionResponse{
		Status: "success",
		Data:   *NewOrganizationCollection(organizations),
	}
}

// NewPaginatedOrganizationCollectionResponse creates a new paginated organization collection response.
// This function wraps a paginated OrganizationCollection in a standard API response format
// with a success status message and includes all pagination metadata.
//
// Parameters:
//   - organizations: Slice of organization domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated OrganizationCollectionResponse
func NewPaginatedOrganizationCollectionResponse(
	organizations []*entities.Organization,
	page, perPage int,
	total int64,
	baseURL string,
) *OrganizationCollectionResponse {
	return &OrganizationCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedOrganizationCollection(organizations, page, perPage, total, baseURL),
	}
}
