// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/utils"
)

// AuthResource represents the authentication resource in API responses.
// This struct contains all the necessary information for a successful authentication,
// including user details and token information for maintaining the user session.
type AuthResource struct {
	// User contains the authenticated user's basic information
	User *AuthUserResource `json:"user"`
	// AccessToken is the JWT token used for API authentication
	AccessToken string `json:"access_token"`
	// RefreshToken is the token used to obtain a new access token when it expires
	RefreshToken string `json:"refresh_token"`
	// ExpiresIn indicates the number of seconds until the access token expires
	ExpiresIn int64 `json:"expires_in"`
	// TokenType specifies the type of authentication token (typically "Bearer")
	TokenType string `json:"token_type"`
}

// AuthUserResource represents a user in auth API responses (simplified version).
// This struct contains only the essential user information needed for authentication
// responses, excluding sensitive data like passwords or detailed profile information.
type AuthUserResource struct {
	// ID is the unique identifier for the user
	ID string `json:"id"`
	// Username is the user's chosen username for login
	Username string `json:"username"`
	// Email is the user's email address
	Email string `json:"email"`
	// Phone is the user's phone number
	Phone string `json:"phone"`
	// Avatar is an optional URL to the user's profile picture
	Avatar *string `json:"avatar,omitempty"`
}

// TokenResource represents a token resource in API responses.
// This struct contains token information without user details, useful for
// token refresh operations or when only token data is needed.
type TokenResource struct {
	// AccessToken is the JWT token used for API authentication
	AccessToken string `json:"access_token"`
	// RefreshToken is the token used to obtain a new access token when it expires
	RefreshToken string `json:"refresh_token"`
	// ExpiresIn indicates the number of seconds until the access token expires
	ExpiresIn int64 `json:"expires_in"`
	// TokenType specifies the type of authentication token (typically "Bearer")
	TokenType string `json:"token_type"`
}

// AuthResourceResponse represents a single auth response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type AuthResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the authentication resource
	Data AuthResource `json:"data"`
}

// TokenResourceResponse represents a single token response.
// This wrapper provides a consistent response structure for token operations
// with response codes and messages.
type TokenResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the token resource
	Data TokenResource `json:"data"`
}

// NewAuthResource creates a new AuthResource from user and token pair.
// This function transforms domain entities into a consistent API response format,
// ensuring all authentication data is properly structured and formatted.
//
// Parameters:
//   - user: The user domain entity
//   - tokenPair: The token pair containing access and refresh tokens
//
// Returns:
//   - A new AuthResource with the provided user and token information
func NewAuthResource(user *entities.User, tokenPair *utils.TokenPair) AuthResource {
	return AuthResource{
		User: &AuthUserResource{
			ID:       user.ID.String(),
			Username: user.UserName,
			Email:    user.Email,
			Phone:    user.Phone,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// NewTokenResource creates a new TokenResource from token pair.
// This function creates a token resource without user information,
// useful for operations that only need to return token data.
//
// Parameters:
//   - tokenPair: The token pair containing access and refresh tokens
//
// Returns:
//   - A new TokenResource with the provided token information
func NewTokenResource(tokenPair *utils.TokenPair) TokenResource {
	return TokenResource{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// Note: NewUserResource is defined in user_responses.go

// NewAuthResourceResponse creates a new AuthResourceResponse.
// This function wraps an AuthResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - user: The user domain entity
//   - tokenPair: The token pair containing access and refresh tokens
//
// Returns:
//   - A new AuthResourceResponse with success status and authentication data
func NewAuthResourceResponse(user *entities.User, tokenPair *utils.TokenPair) AuthResourceResponse {
	return AuthResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Authentication successful",
		Data:            NewAuthResource(user, tokenPair),
	}
}

// NewTokenResourceResponse creates a new TokenResourceResponse.
// This function wraps a TokenResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - tokenPair: The token pair containing access and refresh tokens
//
// Returns:
//   - A new TokenResourceResponse with success status and token data
func NewTokenResourceResponse(tokenPair *utils.TokenPair) TokenResourceResponse {
	return TokenResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Token operation successful",
		Data:            NewTokenResource(tokenPair),
	}
}

// Legacy response types for backward compatibility
// These will be deprecated in favor of the new resource responses

// AuthResponse represents the authentication response (legacy).
// This struct is maintained for backward compatibility with existing clients
// and will be deprecated in future versions in favor of AuthResourceResponse.
type AuthResponse struct {
	// User contains the authenticated user's information
	User *UserResponse `json:"user"`
	// AccessToken is the JWT token used for API authentication
	AccessToken string `json:"access_token"`
	// RefreshToken is the token used to obtain a new access token when it expires
	RefreshToken string `json:"refresh_token"`
	// ExpiresIn indicates the number of seconds until the access token expires
	ExpiresIn int64 `json:"expires_in"`
	// TokenType specifies the type of authentication token (typically "Bearer")
	TokenType string `json:"token_type"`
}

// UserResponse represents a user in API responses (legacy).
// This struct is maintained for backward compatibility and includes
// additional fields not present in the simplified AuthUserResource.
type UserResponse struct {
	// ID is the unique identifier for the user
	ID string `json:"id"`
	// Username is the user's chosen username for login
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
}

// TokenResponse represents a token response (legacy).
// This struct is maintained for backward compatibility and will be
// deprecated in favor of TokenResource.
type TokenResponse struct {
	// AccessToken is the JWT token used for API authentication
	AccessToken string `json:"access_token"`
	// RefreshToken is the token used to obtain a new access token when it expires
	RefreshToken string `json:"refresh_token"`
	// ExpiresIn indicates the number of seconds until the access token expires
	ExpiresIn int64 `json:"expires_in"`
	// TokenType specifies the type of authentication token (typically "Bearer")
	TokenType string `json:"token_type"`
}

// NewAuthResponse creates a new AuthResponse from user and token pair (legacy).
// This function is maintained for backward compatibility and will be
// deprecated in favor of NewAuthResourceResponse.
//
// Parameters:
//   - user: The user domain entity
//   - tokenPair: The token pair containing access and refresh tokens
//
// Returns:
//   - A pointer to the newly created legacy AuthResponse
func NewAuthResponse(user *entities.User, tokenPair *utils.TokenPair) *AuthResponse {
	return &AuthResponse{
		User:         NewUserResponse(user),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// NewTokenResponse creates a new TokenResponse from token pair (legacy).
// This function is maintained for backward compatibility and will be
// deprecated in favor of NewTokenResourceResponse.
//
// Parameters:
//   - tokenPair: The token pair containing access and refresh tokens
//
// Returns:
//   - A pointer to the newly created legacy TokenResponse
func NewTokenResponse(tokenPair *utils.TokenPair) *TokenResponse {
	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// NewUserResponse creates a new UserResponse from user entity (legacy).
// This function is maintained for backward compatibility and will be
// deprecated in favor of the new resource pattern.
//
// Parameters:
//   - user: The user domain entity to convert
//
// Returns:
//   - A pointer to the newly created legacy UserResponse
func NewUserResponse(user *entities.User) *UserResponse {
	var avatar *string
	if user.Avatar != "" {
		avatar = &user.Avatar
	}

	return &UserResponse{
		ID:       user.ID.String(),
		Username: user.UserName,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   avatar,
	}
}

// UserListResponse represents a list of users with pagination (legacy).
// This struct is maintained for backward compatibility and will be
// deprecated in favor of the new collection pattern with proper metadata.
type UserListResponse struct {
	// Users contains the array of user responses
	Users []UserResponse `json:"users"`
	// Total indicates the total number of users across all pages
	Total int64 `json:"total"`
	// Limit specifies the maximum number of users per page
	Limit int `json:"limit"`
	// Page indicates the current page number
	Page int `json:"page"`
}

// NewUserListResponse creates a new UserListResponse from user entities (legacy).
// This function is maintained for backward compatibility and will be
// deprecated in favor of the new collection pattern.
//
// Parameters:
//   - users: Slice of user domain entities to convert
//   - total: Total number of users across all pages
//   - limit: Maximum number of users per page
//   - page: Current page number
//
// Returns:
//   - A pointer to the newly created legacy UserListResponse
func NewUserListResponse(users []*entities.User, total int64, limit, page int) *UserListResponse {
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *NewUserResponse(user)
	}

	return &UserListResponse{
		Users: userResponses,
		Total: total,
		Limit: limit,
		Page:  page,
	}
}
