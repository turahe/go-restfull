package adapters

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresSettingRepository struct {
	repo repository.SettingRepository
}

func NewPostgresSettingRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repository.SettingRepository {
	return &PostgresSettingRepository{
		repo: repository.NewSettingRepository(db, redisClient),
	}
}

func (r *PostgresSettingRepository) Create(ctx context.Context, setting *entities.Setting) error {
	return r.repo.Create(ctx, setting)
}

func (r *PostgresSettingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *PostgresSettingRepository) GetByKey(ctx context.Context, key string) (*entities.Setting, error) {
	return r.repo.GetByKey(ctx, key)
}

func (r *PostgresSettingRepository) GetAll(ctx context.Context) ([]*entities.Setting, error) {
	return r.repo.GetAll(ctx)
}

func (r *PostgresSettingRepository) Update(ctx context.Context, setting *entities.Setting) error {
	return r.repo.Update(ctx, setting)
}

func (r *PostgresSettingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

func (r *PostgresSettingRepository) ExistsByKey(ctx context.Context, key string) (bool, error) {
	return r.repo.ExistsByKey(ctx, key)
}

func (r *PostgresSettingRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}
