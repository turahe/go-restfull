package queries

import (
	"github.com/google/uuid"
)

// GetUserByIDQuery represents a query to get a user by ID
type GetUserByIDQuery struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// GetUserByEmailQuery represents a query to get a user by email
type GetUserByEmailQuery struct {
	Email string `json:"email" validate:"required,email"`
}

// GetUserByUsernameQuery represents a query to get a user by username
type GetUserByUsernameQuery struct {
	Username string `json:"username" validate:"required"`
}

// GetUserByPhoneQuery represents a query to get a user by phone
type GetUserByPhoneQuery struct {
	Phone string `json:"phone" validate:"required"`
}

// ListUsersQuery represents a query to list users with pagination and filters
type ListUsersQuery struct {
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
	Search   *string `json:"search,omitempty"`
	RoleID   *uuid.UUID `json:"role_id,omitempty"`
	Status   *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	SortBy   *string `json:"sort_by,omitempty" validate:"omitempty,oneof=username email created_at updated_at"`
	SortDir  *string `json:"sort_dir,omitempty" validate:"omitempty,oneof=asc desc"`
}

// GetUserRolesQuery represents a query to get user roles
type GetUserRolesQuery struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// GetUserProfileQuery represents a query to get user profile
type GetUserProfileQuery struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// SearchUsersQuery represents a query to search users
type SearchUsersQuery struct {
	Query    string `json:"query" validate:"required,min=3"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}