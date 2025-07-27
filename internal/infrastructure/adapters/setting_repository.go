package adapters

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SettingRepository defines the interface for setting operations
type SettingRepository interface {
	GetSetting(ctx context.Context) (model.Setting, error)
	GetSettingByKey(ctx context.Context, key string) (model.Setting, error)
	SetSetting(ctx context.Context, setting model.Setting) error
	SetModelSetting(ctx context.Context, setting model.Setting) error
	UpdateSetting(ctx context.Context, key string, value string) (model.Setting, error)
	DeleteSetting(ctx context.Context, setting model.Setting) error
}

type PostgresSettingRepository struct {
	repo repository.SettingRepository
}

func NewPostgresSettingRepository(db *pgxpool.Pool, redisClient redis.Cmdable) SettingRepository {
	return &PostgresSettingRepository{
		repo: repository.NewSettingRepository(db, redisClient),
	}
}

func (r *PostgresSettingRepository) GetSetting(ctx context.Context) (model.Setting, error) {
	return r.repo.GetSetting(ctx)
}

func (r *PostgresSettingRepository) GetSettingByKey(ctx context.Context, key string) (model.Setting, error) {
	return r.repo.GetSettingByKey(ctx, key)
}

func (r *PostgresSettingRepository) SetSetting(ctx context.Context, setting model.Setting) error {
	return r.repo.SetSetting(ctx, setting)
}

func (r *PostgresSettingRepository) SetModelSetting(ctx context.Context, setting model.Setting) error {
	return r.repo.SetModelSetting(ctx, setting)
}

func (r *PostgresSettingRepository) UpdateSetting(ctx context.Context, key string, value string) (model.Setting, error) {
	return r.repo.UpdateSetting(ctx, key, value)
}

func (r *PostgresSettingRepository) DeleteSetting(ctx context.Context, setting model.Setting) error {
	return r.repo.DeleteSetting(ctx, setting)
}
