package repositories

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entities.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entities.User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*entities.User, error)

	// GetByPhone retrieves a user by phone
	GetByPhone(ctx context.Context, phone string) (*entities.User, error)

	// GetAll retrieves all users with optional pagination
	GetAll(ctx context.Context, limit, offset int) ([]*entities.User, error)

	// Search searches users by query
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entities.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user exists with the given username
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByPhone checks if a user exists with the given phone
	ExistsByPhone(ctx context.Context, phone string) (bool, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)

	// CountBySearch returns the total number of users matching the search query
	CountBySearch(ctx context.Context, query string) (int64, error)
}
