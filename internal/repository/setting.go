package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"webapi/internal/db/model"
)

type SettingRepository interface {
	GetSetting(ctx context.Context) (model.Setting, error)
	GetSettingByKey(ctx context.Context, key string) (model.Setting, error)
	SetSetting(ctx context.Context, setting model.Setting) error
	SetModelSetting(ctx context.Context, setting model.Setting) error
	UpdateSetting(ctx context.Context, key string, value string) (model.Setting, error)
	DeleteSetting(ctx context.Context, setting model.Setting) error
}

type SettingRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewSettingRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) SettingRepository {
	return &SettingRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}
func (s *SettingRepositoryImpl) GetSetting(ctx context.Context) (model.Setting, error) {
	var settingModel model.Setting
	err := s.pgxPool.QueryRow(ctx, "SELECT id, key, value FROM settings").Scan(
		&settingModel.ID,
		&settingModel.Key,
		&settingModel.Value,
	)
	if err != nil {
		return model.Setting{}, err
	}
	return settingModel, nil
}
func (s *SettingRepositoryImpl) GetSettingByKey(ctx context.Context, key string) (model.Setting, error) {
	var settingModel model.Setting
	err := s.pgxPool.QueryRow(ctx, "SELECT id, key, value FROM settings WHERE key = $1", key).Scan(
		&settingModel.ID,
		&settingModel.Key,
		&settingModel.Value,
	)
	if err != nil {
		return model.Setting{}, err
	}
	return settingModel, nil
}
func (s *SettingRepositoryImpl) SetSetting(ctx context.Context, setting model.Setting) error {
	_, err := s.pgxPool.Exec(ctx, "INSERT INTO settings (key, value) VALUES ($1, $2)", setting.Key, setting.Value)
	if err != nil {
		return err
	}
	return nil
}
func (s *SettingRepositoryImpl) SetModelSetting(ctx context.Context, setting model.Setting) error {
	_, err := s.pgxPool.Exec(ctx, "INSERT INTO settings (id, model_type, model_id, key, value) VALUES ($1, $2, $3, $4, $5)", uuid.New(), setting.ModelType, setting.ModelId, setting.Key, setting.Value)
	if err != nil {
		return err
	}
	return nil
}
func (s *SettingRepositoryImpl) UpdateModelSetting(ctx context.Context, modelId uuid.UUID) (model.Setting, error) {
	var settingModel model.Setting
	err := s.pgxPool.QueryRow(ctx, "UPDATE settings SET model_id = $1 WHERE id = $2 RETURNING id, key, value", modelId, settingModel.ID).Scan(
		&settingModel.ID,
		&settingModel.Key,
		&settingModel.Value,
	)
	if err != nil {
		return model.Setting{}, err
	}
	return settingModel, nil
}

func (s *SettingRepositoryImpl) UpdateSetting(ctx context.Context, key string, value string) (model.Setting, error) {
	var settingModel model.Setting
	err := s.pgxPool.QueryRow(ctx, "UPDATE settings SET value = $2 WHERE id = $1 RETURNING id, key, value", value, key).Scan(
		&settingModel.ID,
		&settingModel.Key,
		&settingModel.Value,
	)
	if err != nil {
		return model.Setting{}, err
	}
	return settingModel, nil
}
func (s *SettingRepositoryImpl) DeleteSetting(ctx context.Context, setting model.Setting) error {
	_, err := s.pgxPool.Exec(ctx, "DELETE FROM settings WHERE key = $1", setting.Key)
	if err != nil {
		return err
	}
	return nil
}
