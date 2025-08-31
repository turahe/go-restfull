// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// UserResource represents a single user in API responses.
// This struct follows the Laravel API Resource pattern for consistent formatting
// and provides a comprehensive view of user data including profile information,
// verification status, associated roles and menus, and audit trail.
type UserResource struct {
	// ID is the unique identifier for the user
	ID string `json:"id"`
	// Username is the user's chosen username for login and display
	Username string `json:"username"`
	// Email is the user's email address
	Email string `json:"email"`
	// Phone is the user's phone number
	Phone string `json:"phone"`
	// Avatar is an optional URL to the user's profile picture
	Avatar *string `json:"avatar,omitempty"`
	// EmailVerifiedAt indicates when the user's email was verified
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	// PhoneVerifiedAt indicates when the user's phone was verified
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`
	// IsEmailVerified indicates whether the user's email has been verified
	IsEmailVerified bool `json:"is_email_verified"`
	// IsPhoneVerified indicates whether the user's phone has been verified
	IsPhoneVerified bool `json:"is_phone_verified"`
	// HasAvatar indicates whether the user has uploaded a profile picture
	HasAvatar bool `json:"has_avatar"`
	// Roles contains the user's assigned roles for access control
	Roles []RoleResource `json:"roles,omitempty"`
	// Menus contains the user's accessible menu items
	Menus []MenuResource `json:"menus,omitempty"`
	// CreatedAt is the timestamp when the user account was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the user account was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is the optional timestamp when the user account was soft-deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// RoleResource is defined in role_responses.go to avoid duplication

// MenuResource represents a menu in user responses.
// This struct provides a simplified view of menu data for user responses,
// including basic menu information and hierarchical structure.
type MenuResource struct {
	// ID is the unique identifier for the menu item
	ID string `json:"id"`
	// Name is the display name of the menu item
	Name string `json:"name"`
	// Slug is the URL-friendly version of the menu item name
	Slug string `json:"slug"`
	// Description is an optional description of the menu item
	Description string `json:"description,omitempty"`
	// ParentID is the optional ID of the parent menu item for hierarchical menus
	ParentID *string `json:"parent_id,omitempty"`
	// Order determines the display order of the menu item
	Order *int64 `json:"order,omitempty"`
	// CreatedAt is the timestamp when the menu item was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the menu item was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// UserCollection represents a collection of users.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type UserCollection struct {
	// Data contains the array of user resources
	Data []UserResource `json:"data"`
	// Meta contains optional collection metadata (pagination, counts, etc.)
	Meta *CollectionMeta `json:"meta,omitempty"`
	// Links contains optional navigation links (first, last, prev, next)
	Links *CollectionLinks `json:"links,omitempty"`
}

// UserResourceResponse represents a single user response with Laravel-style formatting.
// This wrapper provides a consistent response structure with status information
// and follows the standard API response format used throughout the application.
type UserResourceResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the user resource
	Data UserResource `json:"data"`
}

// UserCollectionResponse represents a collection response with Laravel-style formatting.
// This wrapper provides a consistent response structure for collections with status information
// and follows the standard API response format used throughout the application.
type UserCollectionResponse struct {
	// Status indicates the success status of the operation
	Status string `json:"status"`
	// Data contains the user collection
	Data UserCollection `json:"data"`
}

// NewUserResource creates a new UserResource from a User entity.
// This function transforms the domain entity into a consistent API response format,
// handling all optional fields, computed properties, and nested resources.
//
// Parameters:
//   - user: The user domain entity to convert
//
// Returns:
//   - A pointer to the newly created UserResource
func NewUserResource(user *entities.User) *UserResource {
	// Handle optional avatar field
	var avatar *string
	if user.Avatar != "" {
		avatar = &user.Avatar
	}

	// Transform associated roles if available
	var roles []RoleResource
	if user.Roles != nil {
		roles = make([]RoleResource, len(user.Roles))
		for i, role := range user.Roles {
			roles[i] = NewRoleResource(role)
		}
	}

	// Transform associated menus if available
	var menus []MenuResource
	if user.Menus != nil {
		menus = make([]MenuResource, len(user.Menus))
		for i, menu := range user.Menus {
			// Handle optional parent ID for hierarchical menus
			var parentID *string
			if menu.ParentID != nil {
				parentIDStr := menu.ParentID.String()
				parentID = &parentIDStr
			}

			menus[i] = MenuResource{
				ID:          menu.ID.String(),
				Name:        menu.Name,
				Slug:        menu.Slug,
				Description: menu.Description,
				ParentID:    parentID,
				Order:       menu.RecordOrdering,
				CreatedAt:   menu.CreatedAt,
				UpdatedAt:   menu.UpdatedAt,
			}
		}
	}

	return &UserResource{
		ID:              user.ID.String(),
		Username:        user.UserName,
		Email:           user.Email,
		Phone:           user.Phone,
		Avatar:          avatar,
		EmailVerifiedAt: user.EmailVerifiedAt,
		PhoneVerifiedAt: user.PhoneVerifiedAt,
		IsEmailVerified: user.EmailVerifiedAt != nil,
		IsPhoneVerified: user.PhoneVerifiedAt != nil,
		HasAvatar:       user.Avatar != "",
		Roles:           roles,
		Menus:           menus,
	}
}

// NewUserCollection creates a new UserCollection from a slice of User entities.
// This function transforms multiple domain entities into a consistent API response format,
// creating a collection that can be easily serialized to JSON.
//
// Parameters:
//   - users: Slice of user domain entities to convert
//
// Returns:
//   - A pointer to the newly created UserCollection
func NewUserCollection(users []*entities.User) *UserCollection {
	userResources := make([]UserResource, len(users))
	for i, user := range users {
		userResources[i] = *NewUserResource(user)
	}

	return &UserCollection{
		Data: userResources,
	}
}

// NewPaginatedUserCollection creates a new UserCollection with pagination metadata.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - users: Slice of user domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated UserCollection
func NewPaginatedUserCollection(
	users []*entities.User,
	page, perPage int,
	total int64,
	baseURL string,
) *UserCollection {
	collection := NewUserCollection(users)

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

// NewUserResourceResponse creates a new single user response.
// This function wraps a UserResource in a standard API response format
// with a success status message.
//
// Parameters:
//   - user: The user domain entity to convert and wrap
//
// Returns:
//   - A pointer to the newly created UserResourceResponse
func NewUserResourceResponse(user *entities.User) *UserResourceResponse {
	return &UserResourceResponse{
		Status: "success",
		Data:   *NewUserResource(user),
	}
}

// NewUserCollectionResponse creates a new user collection response.
// This function wraps a UserCollection in a standard API response format
// with a success status message.
//
// Parameters:
//   - users: Slice of user domain entities to convert and wrap
//
// Returns:
//   - A pointer to the newly created UserCollectionResponse
func NewUserCollectionResponse(users []*entities.User) *UserCollectionResponse {
	return &UserCollectionResponse{
		Status: "success",
		Data:   *NewUserCollection(users),
	}
}

// NewPaginatedUserCollectionResponse creates a new paginated user collection response.
// This function wraps a paginated UserCollection in a standard API response format
// with a success status message and includes all pagination metadata.
//
// Parameters:
//   - users: Slice of user domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A pointer to the newly created paginated UserCollectionResponse
func NewPaginatedUserCollectionResponse(
	users []*entities.User,
	page, perPage int,
	total int64,
	baseURL string,
) *UserCollectionResponse {
	return &UserCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedUserCollection(users, page, perPage, total, baseURL),
	}
}
