package repository

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type SettingRepository interface {
	Create(ctx context.Context, setting *entities.Setting) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error)
	GetByKey(ctx context.Context, key string) (*entities.Setting, error)
	GetAll(ctx context.Context) ([]*entities.Setting, error)
	Update(ctx context.Context, setting *entities.Setting) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByKey(ctx context.Context, key string) (bool, error)
	Count(ctx context.Context) (int64, error)
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

func (r *SettingRepositoryImpl) Create(ctx context.Context, setting *entities.Setting) error {
	query := `INSERT INTO settings (id, model_type, model_id, key, value, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}

	_, err := r.pgxPool.Exec(ctx, query,
		setting.ID.String(), setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.CreatedAt, setting.UpdatedAt)
	return err
}

func (r *SettingRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error) {
	query := `SELECT id, model_type, model_id, key, value, created_at, updated_at, deleted_at
			  FROM settings WHERE id = $1 AND deleted_at IS NULL`

	var setting entities.Setting
	var modelIDStr *string

	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert model ID string to UUID
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}

	return &setting, nil
}

func (r *SettingRepositoryImpl) GetByKey(ctx context.Context, key string) (*entities.Setting, error) {
	query := `SELECT id, model_type, model_id, key, value, created_at, updated_at, deleted_at
			  FROM settings WHERE key = $1 AND deleted_at IS NULL`

	var setting entities.Setting
	var modelIDStr *string

	err := r.pgxPool.QueryRow(ctx, query, key).Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert model ID string to UUID
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}

	return &setting, nil
}

func (r *SettingRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Setting, error) {
	query := `SELECT id, model_type, model_id, key, value, created_at, updated_at, deleted_at
			  FROM settings WHERE deleted_at IS NULL
			  ORDER BY key ASC`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*entities.Setting
	for rows.Next() {
		setting, err := r.scanSettingRow(rows)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

func (r *SettingRepositoryImpl) Update(ctx context.Context, setting *entities.Setting) error {
	query := `UPDATE settings SET model_type = $1, model_id = $2, key = $3, value = $4, updated_at = $5
			  WHERE id = $6 AND deleted_at IS NULL`

	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}

	_, err := r.pgxPool.Exec(ctx, query, setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.UpdatedAt, setting.ID.String())
	return err
}

func (r *SettingRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE settings SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *SettingRepositoryImpl) ExistsByKey(ctx context.Context, key string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM settings WHERE key = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, key).Scan(&exists)
	return exists, err
}

func (r *SettingRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM settings WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// scanSettingRow is a helper function to scan a setting row from database
func (r *SettingRepositoryImpl) scanSettingRow(rows pgx.Rows) (*entities.Setting, error) {
	var setting entities.Setting
	var modelIDStr *string

	err := rows.Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert model ID string to UUID
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}

	return &setting, nil
}
