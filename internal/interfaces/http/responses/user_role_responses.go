// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// UserRoleResource represents a user-role relationship in API responses.
// This struct provides a comprehensive view of the relationship between users and roles,
// including both the relationship identifiers and optional nested user and role resources.
type UserRoleResource struct {
	// UserID is the unique identifier for the user in the relationship
	UserID string `json:"user_id"`
	// RoleID is the unique identifier for the role in the relationship
	RoleID string `json:"role_id"`
	// User contains optional user information if the user entity is provided
	User *UserResource `json:"user,omitempty"`
	// Role contains optional role information if the role entity is provided
	Role *RoleResource `json:"role,omitempty"`
}

// UserRoleCollection represents a collection of user-role relationships.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type UserRoleCollection struct {
	// Data contains the array of user-role relationship resources
	Data []UserRoleResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// UserRoleResourceResponse represents a single user-role response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type UserRoleResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the user-role relationship resource
	Data UserRoleResource `json:"data"`
}

// UserRoleCollectionResponse represents a collection of user-role relationships response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type UserRoleCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the user-role relationship collection
	Data UserRoleCollection `json:"data"`
}

// RoleUserResource represents a role-user relationship in API responses.
// This struct provides a comprehensive view of the relationship between roles and users,
// including both the relationship identifiers and optional nested role and user resources.
// It's essentially the inverse of UserRoleResource for different query perspectives.
type RoleUserResource struct {
	// RoleID is the unique identifier for the role in the relationship
	RoleID string `json:"role_id"`
	// UserID is the unique identifier for the user in the relationship
	UserID string `json:"user_id"`
	// Role contains optional role information if the role entity is provided
	Role *RoleResource `json:"role,omitempty"`
	// User contains optional user information if the user entity is provided
	User *UserResource `json:"user,omitempty"`
}

// RoleUserCollection represents a collection of role-user relationships.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type RoleUserCollection struct {
	// Data contains the array of role-user relationship resources
	Data []RoleUserResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// RoleUserCollectionResponse represents a collection of role-user relationships response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type RoleUserCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the role-user relationship collection
	Data RoleUserCollection `json:"data"`
}

// NewUserRoleResource creates a new UserRoleResource from user and role entities.
// This function creates a user-role relationship resource with optional nested
// user and role information if the entities are provided.
//
// Parameters:
//   - userID: The unique identifier for the user
//   - roleID: The unique identifier for the role
//   - user: Optional user entity for nested user information
//   - role: Optional role entity for nested role information
//
// Returns:
//   - A new UserRoleResource with the relationship and optional nested data
func NewUserRoleResource(userID, roleID string, user *entities.User, role *entities.Role) UserRoleResource {
	resource := UserRoleResource{
		UserID: userID,
		RoleID: roleID,
	}

	// Add user information if the user entity is provided
	if user != nil {
		resource.User = &UserResource{
			ID:       user.ID.String(),
			Username: user.UserName,
			Email:    user.Email,
			Phone:    user.Phone,
		}
	}

	// Add role information if the role entity is provided
	if role != nil {
		resource.Role = &RoleResource{
			ID:          role.ID.String(),
			Name:        role.Name,
			Slug:        role.Slug,
			Description: role.Description,
			IsActive:    role.IsActive,
		}
	}

	return resource
}

// NewUserRoleCollection creates a new UserRoleCollection.
// This function creates a collection from a slice of user-role relationship resources.
//
// Parameters:
//   - userRoles: Slice of user-role relationship resources
//
// Returns:
//   - A new UserRoleCollection with the provided relationships
func NewUserRoleCollection(userRoles []UserRoleResource) UserRoleCollection {
	return UserRoleCollection{
		Data: userRoles,
	}
}

// NewUserRoleResourceResponse creates a new UserRoleResourceResponse.
// This function wraps a UserRoleResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - userID: The unique identifier for the user
//   - roleID: The unique identifier for the role
//   - user: Optional user entity for nested user information
//   - role: Optional role entity for nested role information
//
// Returns:
//   - A new UserRoleResourceResponse with success status and user-role data
func NewUserRoleResourceResponse(userID, roleID string, user *entities.User, role *entities.Role) UserRoleResourceResponse {
	return UserRoleResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "User role operation successful",
		Data:            NewUserRoleResource(userID, roleID, user, nil),
	}
}

// NewUserRoleCollectionResponse creates a new UserRoleCollectionResponse.
// This function wraps a UserRoleCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - userRoles: Slice of user-role relationship resources
//
// Returns:
//   - A new UserRoleCollectionResponse with success status and user-role collection data
func NewUserRoleCollectionResponse(userRoles []UserRoleResource) UserRoleCollectionResponse {
	return UserRoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "User roles retrieved successfully",
		Data:            NewUserRoleCollection(userRoles),
	}
}

// NewRoleUserResource creates a new RoleUserResource from role and user entities.
// This function creates a role-user relationship resource with optional nested
// role and user information if the entities are provided.
//
// Parameters:
//   - roleID: The unique identifier for the role
//   - userID: The unique identifier for the user
//   - role: Optional role entity for nested role information
//   - user: Optional user entity for nested user information
//
// Returns:
//   - A new RoleUserResource with the relationship and optional nested data
func NewRoleUserResource(roleID, userID string, role *entities.Role, user *entities.User) RoleUserResource {
	resource := RoleUserResource{
		RoleID: roleID,
		UserID: userID,
	}

	// Add role information if the role entity is provided
	if role != nil {
		resource.Role = &RoleResource{
			ID:          role.ID.String(),
			Name:        role.Name,
			Slug:        role.Slug,
			Description: role.Description,
			IsActive:    role.IsActive,
		}
	}

	// Add user information if the user entity is provided
	if user != nil {
		resource.User = &UserResource{
			ID:       user.ID.String(),
			Username: user.UserName,
			Email:    user.Email,
			Phone:    user.Phone,
		}
	}

	return resource
}

// NewRoleUserCollection creates a new RoleUserCollection.
// This function creates a collection from a slice of role-user relationship resources.
//
// Parameters:
//   - roleUsers: Slice of role-user relationship resources
//
// Returns:
//   - A new RoleUserCollection with the provided relationships
func NewRoleUserCollection(roleUsers []RoleUserResource) RoleUserCollection {
	return RoleUserCollection{
		Data: roleUsers,
	}
}

// NewRoleUserCollectionResponse creates a new RoleUserCollectionResponse.
// This function wraps a RoleUserCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - roleUsers: Slice of role-user relationship resources
//
// Returns:
//   - A new RoleUserCollectionResponse with success status and role-user collection data
func NewRoleUserCollectionResponse(roleUsers []RoleUserResource) RoleUserCollectionResponse {
	return RoleUserCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Role users retrieved successfully",
		Data:            NewRoleUserCollection(roleUsers),
	}
}

// NewPaginatedRoleUserCollection creates a new RoleUserCollection with pagination.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - roleUsers: Slice of role-user relationship resources for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated RoleUserCollection with metadata and navigation links
func NewPaginatedRoleUserCollection(roleUsers []RoleUserResource, page, perPage, total int, baseURL string) RoleUserCollection {
	collection := NewRoleUserCollection(roleUsers)

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

// NewPaginatedRoleUserCollectionResponse creates a new RoleUserCollectionResponse with pagination.
// This function wraps a paginated RoleUserCollection in a standard API response format
// with appropriate response codes and success messages, including all pagination metadata.
//
// Parameters:
//   - roleUsers: Slice of role-user relationship resources for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated RoleUserCollectionResponse with success status and pagination data
func NewPaginatedRoleUserCollectionResponse(roleUsers []RoleUserResource, page, perPage, total int, baseURL string) RoleUserCollectionResponse {
	return RoleUserCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Role users retrieved successfully",
		Data:            NewPaginatedRoleUserCollection(roleUsers, page, perPage, total, baseURL),
	}
}
