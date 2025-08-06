package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
)

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone    *string `json:"phone,omitempty"`
}

// ChangePasswordRequest represents the request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UpdateUserProfileRequest represents the request to update user profile
type UpdateUserProfileRequest struct {
	FirstName   string     `json:"first_name" validate:"required,min=1,max=50"`
	LastName    string     `json:"last_name" validate:"required,min=1,max=50"`
	Avatar      *string    `json:"avatar,omitempty"`
	Bio         *string    `json:"bio,omitempty" validate:"omitempty,max=500"`
	Website     *string    `json:"website,omitempty" validate:"omitempty,url,max=255"`
	Location    *string    `json:"location,omitempty" validate:"omitempty,max=100"`
	Gender      *string    `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID              uuid.UUID        `json:"id"`
	Username        string           `json:"username"`
	Email           string           `json:"email"`
	Phone           string           `json:"phone"`
	EmailVerified   bool             `json:"email_verified"`
	PhoneVerified   bool             `json:"phone_verified"`
	Roles           []RoleResponse   `json:"roles,omitempty"`
	Profile         *ProfileResponse `json:"profile,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Version         int              `json:"version"`
}

// RoleResponse represents the role response
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProfileResponse represents the user profile response
type ProfileResponse struct {
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	FullName    string     `json:"full_name"`
	Avatar      string     `json:"avatar,omitempty"`
	Bio         string     `json:"bio,omitempty"`
	Website     string     `json:"website,omitempty"`
	Location    string     `json:"location,omitempty"`
	Gender      string     `json:"gender,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Age         *int       `json:"age,omitempty"`
}

// PaginatedUsersResponse represents a paginated list of users
type PaginatedUsersResponse struct {
	Items      []UserResponse `json:"items"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// NewUserResponse creates a new user response from a user aggregate
func NewUserResponse(user *aggregates.UserAggregate) UserResponse {
	response := UserResponse{
		ID:            user.ID,
		Username:      user.UserName,
		Email:         user.Email.String(),
		Phone:         user.Phone.String(),
		EmailVerified: user.IsEmailVerified(),
		PhoneVerified: user.IsPhoneVerified(),
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		Version:       user.Version,
	}

	// Convert roles
	if len(user.Roles) > 0 {
		response.Roles = make([]RoleResponse, len(user.Roles))
		for i, role := range user.Roles {
			response.Roles[i] = RoleResponse{
				ID:          role.ID,
				Name:        role.Name,
				Description: role.Description,
				CreatedAt:   role.CreatedAt,
			}
		}
	}

	// Convert profile
	if user.Profile != nil {
		response.Profile = &ProfileResponse{
			FirstName:   user.Profile.FirstName,
			LastName:    user.Profile.LastName,
			FullName:    user.Profile.FullName,
			Avatar:      user.Profile.Avatar,
			Bio:         user.Profile.Bio,
			Website:     user.Profile.Website,
			Location:    user.Profile.Location,
			Gender:      string(user.Profile.Gender),
			DateOfBirth: user.Profile.DateOfBirth,
			Age:         user.Profile.GetAge(),
		}
	}

	return response
}

// NewPaginatedUsersResponse creates a new paginated users response
func NewPaginatedUsersResponse(result *handlers.PaginatedResult[*aggregates.UserAggregate]) PaginatedUsersResponse {
	response := PaginatedUsersResponse{
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		Items:      make([]UserResponse, len(result.Items)),
	}

	for i, user := range result.Items {
		response.Items[i] = NewUserResponse(user)
	}

	return response
}

// NewRoleResponse creates a new role response from a role value object
func NewRoleResponse(role valueobjects.Role) RoleResponse {
	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}