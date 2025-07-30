package adapters

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresRoleRepository struct {
	repo repository.RoleRepository
}

func NewPostgresRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.RoleRepository {
	return &PostgresRoleRepository{
		repo: repository.NewRoleRepository(db, redisClient),
	}
}

func (r *PostgresRoleRepository) Create(ctx context.Context, role *entities.Role) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Create(ctx, role)
}

func (r *PostgresRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByID(ctx, id)
}

func (r *PostgresRoleRepository) GetBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetBySlug(ctx, slug)
}

func (r *PostgresRoleRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetAll(ctx, limit, offset)
}

func (r *PostgresRoleRepository) GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetActive(ctx, limit, offset)
}

func (r *PostgresRoleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Search(ctx, query, limit, offset)
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *entities.Role) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Update(ctx, role)
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Delete(ctx, id)
}

func (r *PostgresRoleRepository) Activate(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Activate(ctx, id)
}

func (r *PostgresRoleRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Deactivate(ctx, id)
}

func (r *PostgresRoleRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.ExistsBySlug(ctx, slug)
}

func (r *PostgresRoleRepository) Count(ctx context.Context) (int64, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Count(ctx)
}

func (r *PostgresRoleRepository) CountActive(ctx context.Context) (int64, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.CountActive(ctx)
}
