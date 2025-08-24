package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresSettingRepository struct {
	*BaseTransactionalRepository
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostgresSettingRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.SettingRepository {
	return &PostgresSettingRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		redisClient:                 redisClient,
	}
}

func (r *PostgresSettingRepository) Create(ctx context.Context, setting *entities.Setting) error {
	query := `INSERT INTO settings (id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}
	_, err := r.db.Exec(ctx, query,
		setting.ID.String(), setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.CreatedBy, setting.UpdatedBy, setting.CreatedAt, setting.UpdatedAt)
	return err
}

func (r *PostgresSettingRepository) BatchCreate(ctx context.Context, settings []*entities.Setting) error {
	if len(settings) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO settings (id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	batch := &pgx.Batch{}
	for _, setting := range settings {
		modelIDStr := ""
		if setting.ModelID != nil {
			modelIDStr = setting.ModelID.String()
		}
		batch.Queue(query,
			setting.ID.String(), setting.ModelType, modelIDStr, setting.Key, setting.Value,
			setting.CreatedBy, setting.UpdatedBy, setting.CreatedAt, setting.UpdatedAt)
	}
	br := tx.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresSettingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error) {
	// cache first
	var cached entities.Setting
	if ok, err := cacheGetJSON(ctx, r.redisClient, "settings:id:"+id.String(), &cached); err == nil && ok {
		return &cached, nil
	}

	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE id = $1 AND deleted_at IS NULL`
	var setting entities.Setting
	var modelIDStr *string
	if err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedBy, &setting.UpdatedBy, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt,
	); err != nil {
		return nil, err
	}
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}
	_ = cacheSetJSON(ctx, r.redisClient, "settings:id:"+id.String(), setting, 300000000000)
	return &setting, nil
}

func (r *PostgresSettingRepository) GetByKey(ctx context.Context, key string) (*entities.Setting, error) {
	// cache first
	var cached entities.Setting
	if ok, err := cacheGetJSON(ctx, r.redisClient, "settings:key:"+key, &cached); err == nil && ok {
		return &cached, nil
	}

	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE key = $1 AND deleted_at IS NULL`
	var setting entities.Setting
	var modelIDStr *string
	if err := r.db.QueryRow(ctx, query, key).Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedBy, &setting.UpdatedBy, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt,
	); err != nil {
		return nil, err
	}
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}
	_ = cacheSetJSON(ctx, r.redisClient, "settings:key:"+key, setting, 300000000000)
	return &setting, nil
}

func (r *PostgresSettingRepository) GetAll(ctx context.Context) ([]*entities.Setting, error) {
	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE deleted_at IS NULL ORDER BY key ASC`
	rows, err := r.db.Query(ctx, query)
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

func (r *PostgresSettingRepository) Update(ctx context.Context, setting *entities.Setting) error {
	query := `UPDATE settings SET model_type = $1, model_id = $2, key = $3, value = $4, updated_at = $5, updated_by = $6
			  WHERE id = $7 AND deleted_at IS NULL`
	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}
	_, err := r.db.Exec(ctx, query, setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.UpdatedAt, setting.UpdatedBy, setting.ID.String())
	if err == nil {
		_ = cacheDelete(ctx, r.redisClient, "settings:id:"+setting.ID.String(), "settings:key:"+setting.Key)
	}
	return err
}

func (r *PostgresSettingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE settings SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id.String())
	if err == nil {
		_ = cacheDelete(ctx, r.redisClient, "settings:id:"+id.String())
		_ = cacheDeleteByPattern(ctx, r.redisClient, "settings:key:*")
	}
	return err
}

func (r *PostgresSettingRepository) ExistsByKey(ctx context.Context, key string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM settings WHERE key = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, key).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresSettingRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM settings WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresSettingRepository) scanSettingRow(rows pgx.Rows) (*entities.Setting, error) {
	var setting entities.Setting
	var modelIDStr *string
	if err := rows.Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedBy, &setting.UpdatedBy, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt,
	); err != nil {
		return nil, err
	}
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}
	return &setting, nil
}
