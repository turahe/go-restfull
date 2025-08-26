package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/utils"
)

// AuthResource represents the authentication resource in API responses
type AuthResource struct {
	User         *AuthUserResource `json:"user"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	ExpiresIn    int64             `json:"expires_in"`
	TokenType    string            `json:"token_type"`
}

// AuthUserResource represents a user in auth API responses (simplified version)
type AuthUserResource struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Avatar   *string `json:"avatar,omitempty"`
}

// TokenResource represents a token resource in API responses
type TokenResource struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// AuthResourceResponse represents a single auth response
type AuthResourceResponse struct {
	ResponseCode    int          `json:"response_code"`
	ResponseMessage string       `json:"response_message"`
	Data            AuthResource `json:"data"`
}

// TokenResourceResponse represents a single token response
type TokenResourceResponse struct {
	ResponseCode    int           `json:"response_code"`
	ResponseMessage string        `json:"response_message"`
	Data            TokenResource `json:"data"`
}

// NewAuthResource creates a new AuthResource from user and token pair
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

// NewTokenResource creates a new TokenResource from token pair
func NewTokenResource(tokenPair *utils.TokenPair) TokenResource {
	return TokenResource{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// Note: NewUserResource is defined in user_responses.go

// NewAuthResourceResponse creates a new AuthResourceResponse
func NewAuthResourceResponse(user *entities.User, tokenPair *utils.TokenPair) AuthResourceResponse {
	return AuthResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Authentication successful",
		Data:            NewAuthResource(user, tokenPair),
	}
}

// NewTokenResourceResponse creates a new TokenResourceResponse
func NewTokenResourceResponse(tokenPair *utils.TokenPair) TokenResourceResponse {
	return TokenResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Token operation successful",
		Data:            NewTokenResource(tokenPair),
	}
}

// Legacy response types for backward compatibility
// These will be deprecated in favor of the new resource responses

// AuthResponse represents the authentication response (legacy)
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
	TokenType    string        `json:"token_type"`
}

// UserResponse represents a user in API responses (legacy)
type UserResponse struct {
	ID              string     `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	Avatar          *string    `json:"avatar,omitempty"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`
}

// TokenResponse represents a token response (legacy)
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// NewAuthResponse creates a new AuthResponse from user and token pair (legacy)
func NewAuthResponse(user *entities.User, tokenPair *utils.TokenPair) *AuthResponse {
	return &AuthResponse{
		User:         NewUserResponse(user),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// NewTokenResponse creates a new TokenResponse from token pair (legacy)
func NewTokenResponse(tokenPair *utils.TokenPair) *TokenResponse {
	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// NewUserResponse creates a new UserResponse from user entity (legacy)
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

// UserListResponse represents a list of users with pagination (legacy)
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
	Limit int            `json:"limit"`
	Page  int            `json:"page"`
}

// NewUserListResponse creates a new UserListResponse from user entities (legacy)
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
