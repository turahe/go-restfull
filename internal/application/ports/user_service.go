package ports

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// UserService defines the application service interface for user operations
type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *entities.User) (*entities.User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)

	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)

	// GetUserByPhone retrieves a user by phone
	GetUserByPhone(ctx context.Context, phone string) (*entities.User, error)

	// GetAllUsers retrieves all users with pagination
	GetAllUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)

	// SearchUsers searches users by query
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error)

	// GetUsersWithPagination retrieves users with pagination and returns total count
	GetUsersWithPagination(ctx context.Context, page, perPage int, search string) ([]*entities.User, int64, error)

	// GetUsersCount returns total count of users (for pagination)
	GetUsersCount(ctx context.Context, search string) (int64, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, user *entities.User) (*entities.User, error)

	// DeleteUser soft deletes a user
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// ChangePassword changes user password
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error

	// VerifyEmail verifies user email
	VerifyEmail(ctx context.Context, id uuid.UUID) error

	// VerifyPhone verifies user phone
	VerifyPhone(ctx context.Context, id uuid.UUID) error

	// AuthenticateUser authenticates user login
	AuthenticateUser(ctx context.Context, username, password string) (*entities.User, error)
}
