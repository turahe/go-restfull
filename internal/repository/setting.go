// Package repository provides data access layer implementations for the application.
// This file contains the SettingRepository interface and its PostgreSQL implementation
// with Redis caching support.
package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SettingRepository defines the contract for setting data operations.
// It provides methods for CRUD operations on settings with support for
// polymorphic relationships through model_type and model_id fields.
type SettingRepository interface {
	// Create persists a new setting to the database
	Create(ctx context.Context, setting *entities.Setting) error

	// BatchCreate persists multiple settings to the database in a single transaction
	// for better performance when inserting large numbers of settings
	BatchCreate(ctx context.Context, settings []*entities.Setting) error

	// GetByID retrieves a setting by its unique identifier
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error)

	// GetByKey retrieves a setting by its key name
	GetByKey(ctx context.Context, key string) (*entities.Setting, error)

	// GetAll retrieves all non-deleted settings ordered by key
	GetAll(ctx context.Context) ([]*entities.Setting, error)

	// Update modifies an existing setting in the database
	Update(ctx context.Context, setting *entities.Setting) error

	// Delete performs a soft delete by setting deleted_at timestamp
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByKey checks if a setting with the given key exists
	ExistsByKey(ctx context.Context, key string) (bool, error)

	// Count returns the total number of non-deleted settings
	Count(ctx context.Context) (int64, error)
}

// SettingRepositoryImpl provides the PostgreSQL implementation of SettingRepository.
// It uses pgx for database operations and includes Redis for potential caching.
type SettingRepositoryImpl struct {
	pgxPool     *pgxpool.Pool // PostgreSQL connection pool
	redisClient redis.Cmdable // Redis client for caching operations
}

// NewSettingRepository creates a new instance of SettingRepositoryImpl
// with the provided database and cache dependencies.
func NewSettingRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) SettingRepository {
	return &SettingRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

// Create persists a new setting to the database.
// The method handles the conversion of ModelID UUID to string for database storage.
func (r *SettingRepositoryImpl) Create(ctx context.Context, setting *entities.Setting) error {
	query := `INSERT INTO settings (id, model_type, model_id, key, value, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	// Convert ModelID UUID to string for database storage
	// Empty string is used when ModelID is nil
	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}

	_, err := r.pgxPool.Exec(ctx, query,
		setting.ID.String(), setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.CreatedBy, setting.UpdatedBy, setting.CreatedAt, setting.UpdatedAt)
	return err
}

// BatchCreate persists multiple settings to the database in a single transaction.
// This method is more efficient than calling Create multiple times for bulk operations.
// It uses a database transaction to ensure all settings are inserted atomically.
func (r *SettingRepositoryImpl) BatchCreate(ctx context.Context, settings []*entities.Setting) error {
	if len(settings) == 0 {
		return nil // No settings to insert
	}

	// Begin a transaction for atomic batch insertion
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	// Prepare the batch insert statement
	query := `INSERT INTO settings (id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Create a batch for efficient insertion
	batch := &pgx.Batch{}

	for _, setting := range settings {
		// Convert ModelID UUID to string for database storage
		// Empty string is used when ModelID is nil
		modelIDStr := ""
		if setting.ModelID != nil {
			modelIDStr = setting.ModelID.String()
		}

		batch.Queue(query,
			setting.ID.String(), setting.ModelType, modelIDStr, setting.Key, setting.Value,
			setting.CreatedBy, setting.UpdatedBy, setting.CreatedAt, setting.UpdatedAt)
	}

	// Execute the batch
	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Check for any errors during batch execution
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

// GetByID retrieves a setting by its unique identifier.
// Returns nil and error if the setting is not found or has been deleted.
func (r *SettingRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error) {
	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE id = $1 AND deleted_at IS NULL`

	var setting entities.Setting
	var modelIDStr *string

	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedBy, &setting.UpdatedBy, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert model ID string back to UUID if it exists
	// This handles the polymorphic relationship storage format
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}

	return &setting, nil
}

// GetByKey retrieves a setting by its key name.
// Returns nil and error if the setting is not found or has been deleted.
func (r *SettingRepositoryImpl) GetByKey(ctx context.Context, key string) (*entities.Setting, error) {
	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE key = $1 AND deleted_at IS NULL`

	var setting entities.Setting
	var modelIDStr *string

	err := r.pgxPool.QueryRow(ctx, query, key).Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedBy, &setting.UpdatedBy, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert model ID string back to UUID if it exists
	// This handles the polymorphic relationship storage format
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}

	return &setting, nil
}

// GetAll retrieves all non-deleted settings from the database.
// Results are ordered alphabetically by key for consistent output.
func (r *SettingRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Setting, error) {
	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
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

// Update modifies an existing setting in the database.
// Only non-deleted settings can be updated.
func (r *SettingRepositoryImpl) Update(ctx context.Context, setting *entities.Setting) error {
	query := `UPDATE settings SET model_type = $1, model_id = $2, key = $3, value = $4, updated_at = $5, updated_by = $6
			  WHERE id = $7 AND deleted_at IS NULL`

	// Convert ModelID UUID to string for database storage
	// Empty string is used when ModelID is nil
	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}

	_, err := r.pgxPool.Exec(ctx, query, setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.UpdatedAt, setting.UpdatedBy, setting.ID.String())
	return err
}

// Delete performs a soft delete by setting the deleted_at timestamp.
// The record remains in the database but is excluded from normal queries.
func (r *SettingRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE settings SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

// ExistsByKey checks if a setting with the given key exists and is not deleted.
// Returns true if the setting exists, false otherwise.
func (r *SettingRepositoryImpl) ExistsByKey(ctx context.Context, key string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM settings WHERE key = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, key).Scan(&exists)
	return exists, err
}

// Count returns the total number of non-deleted settings in the database.
func (r *SettingRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM settings WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// scanSettingRow is a helper function to scan a setting row from database result set.
// It handles the conversion of model_id from string back to UUID for polymorphic relationships.
// This method is used by GetAll to process multiple rows efficiently.
func (r *SettingRepositoryImpl) scanSettingRow(rows pgx.Rows) (*entities.Setting, error) {
	var setting entities.Setting
	var modelIDStr *string

	err := rows.Scan(
		&setting.ID, &setting.ModelType, &modelIDStr, &setting.Key, &setting.Value,
		&setting.CreatedBy, &setting.UpdatedBy, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert model ID string back to UUID if it exists
	// This handles the polymorphic relationship storage format
	if modelIDStr != nil {
		if modelID, err := uuid.Parse(*modelIDStr); err == nil {
			setting.ModelID = &modelID
		}
	}

	return &setting, nil
}
