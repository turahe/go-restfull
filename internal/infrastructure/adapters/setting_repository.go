// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
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

// PostgresSettingRepository implements the SettingRepository interface using PostgreSQL as the primary
// data store and Redis for caching. This repository provides CRUD operations for setting entities
// with support for polymorphic relationships, batch operations, soft deletes, and intelligent caching.
// It embeds the BaseTransactionalRepository to inherit transaction management capabilities.
type PostgresSettingRepository struct {
	// BaseTransactionalRepository provides transaction management functionality
	*BaseTransactionalRepository
	// db holds the PostgreSQL connection pool for database operations
	db *pgxpool.Pool
	// redisClient holds the Redis client for caching operations
	redisClient redis.Cmdable
}

// NewPostgresSettingRepository creates a new PostgreSQL setting repository instance.
// This factory function initializes the repository with database and Redis connections,
// and sets up the base transactional repository for transaction management.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations
//   - redisClient: The Redis client for caching operations
//
// Returns:
//   - repositories.SettingRepository: A new setting repository instance
func NewPostgresSettingRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.SettingRepository {
	return &PostgresSettingRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		redisClient:                 redisClient,
	}
}

// Create persists a new setting entity to the database.
// This method inserts a new setting record with all required fields including
// audit information (created_by, updated_by, timestamps) and polymorphic relationships.
// The method handles optional model_id conversion from UUID to string for storage.
//
// Parameters:
//   - ctx: Context for the database operation
//   - setting: The setting entity to create
//
// Returns:
//   - error: Any error that occurred during the database operation
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

// BatchCreate persists multiple setting entities to the database in a single transaction.
// This method uses PostgreSQL batch operations for improved performance when creating
// multiple settings simultaneously. It ensures atomicity - if any setting fails to create,
// all changes are rolled back.
//
// Parameters:
//   - ctx: Context for the database operation
//   - settings: Slice of setting entities to create
//
// Returns:
//   - error: Any error that occurred during the batch creation operation
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

// GetByID retrieves a setting entity by its unique identifier.
// This method implements intelligent caching - it first checks Redis cache before
// querying the database. If found in cache, it returns immediately. If not found,
// it queries the database and caches the result for future requests. The cache
// has a TTL of 300 seconds (5 minutes) for optimal performance.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the setting to retrieve
//
// Returns:
//   - *entities.Setting: The found setting entity, or nil if not found
//   - error: Any error that occurred during the database operation
func (r *PostgresSettingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error) {
	// Check cache first for improved performance
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
	// Parse model_id string back to UUID if it exists
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}
	// Cache the result for future requests (5 minute TTL)
	_ = cacheSetJSON(ctx, r.redisClient, "settings:id:"+id.String(), setting, 300000000000)
	return &setting, nil
}

// GetByKey retrieves a setting entity by its key identifier.
// This method implements intelligent caching similar to GetByID, checking Redis cache
// first before querying the database. It's useful for retrieving settings by their
// human-readable key names rather than UUIDs.
//
// Parameters:
//   - ctx: Context for the database operation
//   - key: The key identifier of the setting to retrieve
//
// Returns:
//   - *entities.Setting: The found setting entity, or nil if not found
//   - error: Any error that occurred during the database operation
func (r *PostgresSettingRepository) GetByKey(ctx context.Context, key string) (*entities.Setting, error) {
	// Check cache first for improved performance
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
	// Parse model_id string back to UUID if it exists
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}
	// Cache the result for future requests (5 minute TTL)
	_ = cacheSetJSON(ctx, r.redisClient, "settings:key:"+key, setting, 300000000000)
	return &setting, nil
}

// GetAll retrieves all active setting entities from the database.
// This method returns settings ordered by key name in ascending order for consistent
// and predictable results. It excludes soft-deleted records and uses the helper
// method scanSettingRow for consistent row parsing.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - []*entities.Setting: List of all active setting entities
//   - error: Any error that occurred during the database operation
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

// Update modifies an existing setting entity in the database.
// This method updates the setting's model association, key, value, and audit fields
// while preserving the original ID and creation information. It only updates
// non-deleted settings and automatically invalidates related cache entries
// to maintain cache consistency.
//
// Parameters:
//   - ctx: Context for the database operation
//   - setting: The setting entity with updated values
//
// Returns:
//   - error: Any error that occurred during the database operation
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
		// Invalidate related cache entries to maintain consistency
		_ = cacheDelete(ctx, r.redisClient, "settings:id:"+setting.ID.String(), "settings:key:"+setting.Key)
	}
	return err
}

// Delete performs a soft delete of a setting entity.
// This method marks the setting as deleted by setting the deleted_at timestamp
// rather than physically removing the record. It also invalidates related cache
// entries to maintain cache consistency across the system.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the setting to delete
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresSettingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE settings SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id.String())
	if err == nil {
		// Invalidate related cache entries to maintain consistency
		_ = cacheDelete(ctx, r.redisClient, "settings:id:"+id.String())
		_ = cacheDeleteByPattern(ctx, r.redisClient, "settings:key:*")
	}
	return err
}

// ExistsByKey checks if a setting with the specified key already exists.
// This method is useful for validation purposes, ensuring setting keys are unique
// before creation. It performs a soft-delete aware check.
//
// Parameters:
//   - ctx: Context for the database operation
//   - key: The key to check for existence
//
// Returns:
//   - bool: True if a setting with the key exists, false otherwise
//   - error: Any error that occurred during the database operation
func (r *PostgresSettingRepository) ExistsByKey(ctx context.Context, key string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM settings WHERE key = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, key).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// Count returns the total number of active settings in the system.
// This method excludes soft-deleted records and is useful for system statistics
// and administrative purposes.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - int64: The total count of active settings
//   - error: Any error that occurred during the database operation
func (r *PostgresSettingRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM settings WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// scanSettingRow is a helper method that consistently parses setting rows from database results.
// This method handles the common pattern of scanning setting fields and parsing the optional
// model_id from string back to UUID. It's used by GetAll and other methods that process
// multiple rows to ensure consistent parsing logic.
//
// Parameters:
//   - rows: The database rows to scan from
//
// Returns:
//   - *entities.Setting: The parsed setting entity
//   - error: Any error that occurred during row parsing
func (r *PostgresSettingRepository) scanSettingRow(rows pgx.Rows) (*entities.Setting, error) {
	var setting entities.Setting
	var modelIDStr *string
	if err := rows.Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedBy, &setting.UpdatedBy, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt,
	); err != nil {
		return nil, err
	}
	// Parse model_id string back to UUID if it exists
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}
	return &setting, nil
}
