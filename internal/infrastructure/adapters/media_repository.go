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

type PostgresMediaRepository struct {
	repo repository.MediaRepository
}

func NewPostgresMediaRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MediaRepository {
	return &PostgresMediaRepository{
		repo: repository.NewMediaRepository(db, redisClient),
	}
}

func (r *PostgresMediaRepository) Create(ctx context.Context, media *entities.Media) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Create(ctx, media)
}

func (r *PostgresMediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByID(ctx, id)
}

func (r *PostgresMediaRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByUserID(ctx, userID, limit, offset)
}

func (r *PostgresMediaRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetAll(ctx, limit, offset)
}

func (r *PostgresMediaRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Search(ctx, query, limit, offset)
}

func (r *PostgresMediaRepository) Update(ctx context.Context, media *entities.Media) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Update(ctx, media)
}

func (r *PostgresMediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Delete(ctx, id)
}

func (r *PostgresMediaRepository) Count(ctx context.Context) (int64, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Count(ctx)
}

func (r *PostgresMediaRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.CountByUserID(ctx, userID)
}
