package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
)

// UserRepository defines the contract for user aggregate persistence
type UserRepository interface {
	// Command operations
	Save(ctx context.Context, user *aggregates.UserAggregate) error
	Delete(ctx context.Context, userID uuid.UUID) error

	// Query operations
	FindByID(ctx context.Context, userID uuid.UUID) (*aggregates.UserAggregate, error)
	FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error)
	FindByUsername(ctx context.Context, username string) (*aggregates.UserAggregate, error)
	FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error)
	
	// List and search operations
	FindAll(ctx context.Context, query queries.ListUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error)
	Search(ctx context.Context, query queries.SearchUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error)
	
	// Existence checks
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	
	// Count operations
	Count(ctx context.Context) (int64, error)
	CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
}
