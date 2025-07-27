package responses

import (
	"time"

	"webapi/internal/domain/entities"
	"webapi/internal/helper/utils"
)

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
	TokenType    string        `json:"token_type"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID              string     `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// NewAuthResponse creates a new AuthResponse from user and token pair
func NewAuthResponse(user *entities.User, tokenPair *utils.TokenPair) *AuthResponse {
	return &AuthResponse{
		User:         NewUserResponse(user),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// NewTokenResponse creates a new TokenResponse from token pair
func NewTokenResponse(tokenPair *utils.TokenPair) *TokenResponse {
	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// NewUserResponse creates a new UserResponse from user entity
func NewUserResponse(user *entities.User) *UserResponse {
	return &UserResponse{
		ID:       user.ID.String(),
		Username: user.UserName,
		Email:    user.Email,
		Phone:    user.Phone,
	}
}

// UserListResponse represents a list of users with pagination
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
	Limit int            `json:"limit"`
	Page  int            `json:"page"`
}

// NewUserListResponse creates a new UserListResponse from user entities
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
