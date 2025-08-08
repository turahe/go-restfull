package ports

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/utils"
)

// AuthService defines the application service interface for authentication operations
type AuthService interface {
	// RegisterUser registers a new user and returns authentication tokens
	RegisterUser(ctx context.Context, username, email, phone, password string) (*utils.TokenPair, *entities.User, error)

	// LoginUser authenticates a user and returns authentication tokens
	LoginUser(ctx context.Context, identity, password string) (*utils.TokenPair, *entities.User, error)

	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*utils.TokenPair, error)

	// LogoutUser invalidates the user's tokens
	LogoutUser(ctx context.Context, userID string) error

	// ForgetPassword sends a password reset email with OTP
	ForgetPassword(ctx context.Context, identifier string) error

	// ResetPassword resets password using OTP
	ResetPassword(ctx context.Context, email, otp, newPassword string) error

	// ValidateToken validates a JWT token and returns user claims
	ValidateToken(ctx context.Context, token string) (*utils.TokenClaims, error)
}
