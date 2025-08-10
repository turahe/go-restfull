package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresUserRepository is an adapter that implements the domain UserRepository interface
// by delegating to the concrete repository implementation
type PostgresUserRepository struct {
	*BaseTransactionalRepository
	repo repositories.UserRepository
}

// NewPostgresUserRepository creates a new PostgreSQL user repository adapter
func NewPostgresUserRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.UserRepository {
	return &PostgresUserRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		repo:                        repository.NewUserRepository(db, redisClient),
	}
}

// Save delegates to the underlying repository implementation
func (r *PostgresUserRepository) Save(ctx context.Context, user *aggregates.UserAggregate) error {
	return r.repo.Save(ctx, user)
}

// Delete delegates to the underlying repository implementation
func (r *PostgresUserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.repo.Delete(ctx, userID)
}

// FindByID delegates to the underlying repository implementation
func (r *PostgresUserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*aggregates.UserAggregate, error) {
	return r.repo.FindByID(ctx, userID)
}

// FindByEmail delegates to the underlying repository implementation
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error) {
	return r.repo.FindByEmail(ctx, email)
}

// FindByUsername delegates to the underlying repository implementation
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*aggregates.UserAggregate, error) {
	return r.repo.FindByUsername(ctx, username)
}

// FindByPhone delegates to the underlying repository implementation
func (r *PostgresUserRepository) FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error) {
	return r.repo.FindByPhone(ctx, phone)
}

// FindAll delegates to the underlying repository implementation
func (r *PostgresUserRepository) FindAll(ctx context.Context, query queries.ListUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	return r.repo.FindAll(ctx, query)
}

// Search delegates to the underlying repository implementation
func (r *PostgresUserRepository) Search(ctx context.Context, query queries.SearchUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	return r.repo.Search(ctx, query)
}

// ExistsByEmail delegates to the underlying repository implementation
func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.repo.ExistsByEmail(ctx, email)
}

// ExistsByUsername delegates to the underlying repository implementation
func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.repo.ExistsByUsername(ctx, username)
}

// ExistsByPhone delegates to the underlying repository implementation
func (r *PostgresUserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	return r.repo.ExistsByPhone(ctx, phone)
}

// Count delegates to the underlying repository implementation
func (r *PostgresUserRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

// CountByRole delegates to the underlying repository implementation
func (r *PostgresUserRepository) CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return r.repo.CountByRole(ctx, roleID)
}
