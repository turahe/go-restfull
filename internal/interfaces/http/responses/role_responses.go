// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// RoleResource represents a role in API responses.
// This struct provides a comprehensive view of role data including basic information,
// status flags, audit trail, and computed properties for easy status checking.
// It follows the Laravel API Resource pattern for consistent formatting.
type RoleResource struct {
	// ID is the unique identifier for the role
	ID string `json:"id"`
	// Name is the display name of the role
	Name string `json:"name"`
	// Slug is the URL-friendly version of the role name
	Slug string `json:"slug"`
	// Description is an optional description of the role's purpose
	Description string `json:"description,omitempty"`
	// IsActive indicates whether the role is currently active and usable
	IsActive bool `json:"is_active"`
	// CreatedBy is the ID of the user who created the role
	CreatedBy string `json:"created_by"`
	// UpdatedBy is the ID of the user who last updated the role
	UpdatedBy string `json:"updated_by"`
	// DeletedBy is the optional ID of the user who deleted the role
	DeletedBy *string `json:"deleted_by,omitempty"`
	// CreatedAt is the timestamp when the role was created
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the role was last updated
	UpdatedAt string `json:"updated_at"`
	// DeletedAt is the optional timestamp when the role was soft-deleted
	DeletedAt *string `json:"deleted_at,omitempty"`

	// Computed fields for easy status checking
	// IsDeleted indicates whether the role has been soft-deleted
	IsDeleted bool `json:"is_deleted"`
	// IsActiveRole indicates whether the role is currently active and usable
	IsActiveRole bool `json:"is_active_role"`
}

// RoleCollection represents a collection of roles.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type RoleCollection struct {
	// Data contains the array of role resources
	Data []RoleResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// RoleResourceResponse represents a single role response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type RoleResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the role resource
	Data RoleResource `json:"data"`
}

// RoleCollectionResponse represents a collection of roles response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type RoleCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the role collection
	Data RoleCollection `json:"data"`
}

// NewRoleResource creates a new RoleResource from role entity.
// This function transforms the domain entity into a consistent API response format,
// handling all optional fields and computed properties appropriately.
//
// Parameters:
//   - role: The role domain entity to convert
//
// Returns:
//   - A new RoleResource with all fields properly formatted
func NewRoleResource(role *entities.Role) RoleResource {
	resource := RoleResource{
		ID:          role.ID.String(),
		Name:        role.Name,
		Slug:        role.Slug,
		Description: role.Description,
		IsActive:    role.IsActive,
		CreatedBy:   role.CreatedBy.String(),
		UpdatedBy:   role.UpdatedBy.String(),
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),

		// Set computed fields based on role state
		IsDeleted:    role.IsDeleted(),
		IsActiveRole: role.IsActiveRole(),
	}

	// Handle optional soft deletion information
	if role.DeletedBy != nil {
		deletedBy := role.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if role.DeletedAt != nil {
		deletedAt := role.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	return resource
}

// NewRoleResourceResponse creates a new RoleResourceResponse.
// This function wraps a RoleResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - role: The role domain entity to convert and wrap
//
// Returns:
//   - A new RoleResourceResponse with success status and role data
func NewRoleResourceResponse(role *entities.Role) RoleResourceResponse {
	return RoleResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Role retrieved successfully",
		Data:            NewRoleResource(role),
	}
}

// NewRoleCollection creates a new RoleCollection.
// This function transforms multiple role domain entities into a consistent
// API response format, creating a collection that can be easily serialized to JSON.
//
// Parameters:
//   - roles: Slice of role domain entities to convert
//
// Returns:
//   - A new RoleCollection with all roles properly formatted
func NewRoleCollection(roles []*entities.Role) RoleCollection {
	roleResources := make([]RoleResource, len(roles))
	for i, role := range roles {
		roleResources[i] = NewRoleResource(role)
	}

	return RoleCollection{
		Data: roleResources,
	}
}

// NewPaginatedRoleCollection creates a new RoleCollection with pagination.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - roles: Slice of role domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated RoleCollection with metadata and navigation links
func NewPaginatedRoleCollection(roles []*entities.Role, page, perPage, total int, baseURL string) RoleCollection {
	collection := NewRoleCollection(roles)

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

// NewRoleCollectionResponse creates a new RoleCollectionResponse.
// This function wraps a RoleCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - roles: Slice of role domain entities to convert and wrap
//
// Returns:
//   - A new RoleCollectionResponse with success status and role collection data
func NewRoleCollectionResponse(roles []*entities.Role) RoleCollectionResponse {
	return RoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Roles retrieved successfully",
		Data:            NewRoleCollection(roles),
	}
}

// NewPaginatedRoleCollectionResponse creates a new RoleCollectionResponse with pagination.
// This function wraps a paginated RoleCollection in a standard API response format
// with appropriate response codes and success messages, including all pagination metadata.
//
// Parameters:
//   - roles: Slice of role domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated RoleCollectionResponse with success status and pagination data
func NewPaginatedRoleCollectionResponse(roles []*entities.Role, page, perPage, total int, baseURL string) RoleCollectionResponse {
	return RoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Roles retrieved successfully",
		Data:            NewPaginatedRoleCollection(roles, page, perPage, total, baseURL),
	}
}
