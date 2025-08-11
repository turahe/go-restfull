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
//
// The interface supports:
// - Basic CRUD operations (Create, Read, Update, Delete)
// - Batch operations for improved performance
// - Polymorphic relationships via model_type and model_id
// - Soft delete functionality
// - Existence and counting operations
type SettingRepository interface {
	// Create persists a new setting to the database.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - setting: Pointer to the setting entity to persist
	//
	// Returns:
	//   - error: Database operation error if any
	Create(ctx context.Context, setting *entities.Setting) error

	// BatchCreate persists multiple settings to the database in a single transaction
	// for better performance when inserting large numbers of settings.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - settings: Slice of setting entities to persist
	//
	// Returns:
	//   - error: Database operation error if any
	//
	// Note: This method is more efficient than calling Create multiple times
	// as it uses a single database transaction.
	BatchCreate(ctx context.Context, settings []*entities.Setting) error

	// GetByID retrieves a setting by its unique identifier.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - id: UUID of the setting to retrieve
	//
	// Returns:
	//   - *entities.Setting: Pointer to the found setting entity, or nil if not found
	//   - error: Database operation error if any
	//
	// Note: Only non-deleted settings are returned.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error)

	// GetByKey retrieves a setting by its key name.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - key: String key of the setting to retrieve
	//
	// Returns:
	//   - *entities.Setting: Pointer to the found setting entity, or nil if not found
	//   - error: Database operation error if any
	//
	// Note: Only non-deleted settings are returned.
	GetByKey(ctx context.Context, key string) (*entities.Setting, error)

	// GetAll retrieves all non-deleted settings ordered by key.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//
	// Returns:
	//   - []*entities.Setting: Slice of all non-deleted setting entities
	//   - error: Database operation error if any
	//
	// Note: Results are ordered alphabetically by key for consistent output.
	GetAll(ctx context.Context) ([]*entities.Setting, error)

	// Update modifies an existing setting in the database.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - setting: Pointer to the setting entity with updated values
	//
	// Returns:
	//   - error: Database operation error if any
	//
	// Note: Only non-deleted settings can be updated.
	Update(ctx context.Context, setting *entities.Setting) error

	// Delete performs a soft delete by setting deleted_at timestamp.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - id: UUID of the setting to delete
	//
	// Returns:
	//   - error: Database operation error if any
	//
	// Note: The record remains in the database but is excluded from normal queries.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByKey checks if a setting with the given key exists.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - key: String key to check for existence
	//
	// Returns:
	//   - bool: true if the setting exists and is not deleted, false otherwise
	//   - error: Database operation error if any
	ExistsByKey(ctx context.Context, key string) (bool, error)

	// Count returns the total number of non-deleted settings.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//
	// Returns:
	//   - int64: Total count of non-deleted settings
	//   - error: Database operation error if any
	Count(ctx context.Context) (int64, error)
}

// SettingRepositoryImpl provides the PostgreSQL implementation of SettingRepository.
// It uses pgx for database operations and includes Redis for potential caching.
//
// The implementation features:
// - Connection pooling via pgxpool for efficient database connections
// - Redis integration for future caching capabilities
// - Proper handling of polymorphic relationships
// - Soft delete support
// - Batch operations for performance optimization
type SettingRepositoryImpl struct {
	pgxPool     *pgxpool.Pool // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable // Redis client for potential caching operations
}

// NewSettingRepository creates a new instance of SettingRepositoryImpl
// with the provided database and cache dependencies.
//
// Parameters:
//   - pgxPool: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - SettingRepository: Interface implementation ready for use
func NewSettingRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) SettingRepository {
	return &SettingRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

// Create persists a new setting to the database.
// The method handles the conversion of ModelID UUID to string for database storage.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - setting: Pointer to the setting entity to persist
//
// Returns:
//   - error: Database operation error if any
//
// Implementation details:
// - Converts ModelID UUID to string for polymorphic relationship storage
// - Uses empty string when ModelID is nil
// - Executes INSERT query with all required fields
func (r *SettingRepositoryImpl) Create(ctx context.Context, setting *entities.Setting) error {
	// SQL query to insert a new setting record
	// Fields: id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at
	query := `INSERT INTO settings (id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Convert ModelID UUID to string for database storage
	// Empty string is used when ModelID is nil to handle polymorphic relationships
	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}

	// Execute the INSERT query with all setting fields
	_, err := r.pgxPool.Exec(ctx, query,
		setting.ID.String(), setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.CreatedBy, setting.UpdatedBy, setting.CreatedAt, setting.UpdatedAt)
	return err
}

// BatchCreate persists multiple settings to the database in a single transaction.
// This method is more efficient than calling Create multiple times for bulk operations.
// It uses a database transaction to ensure all settings are inserted atomically.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - settings: Slice of setting entities to persist
//
// Returns:
//   - error: Database operation error if any
//
// Implementation details:
// - Uses database transaction for atomicity
// - Implements batch processing for performance
// - Handles ModelID conversion for each setting
// - Rolls back on any error during batch execution
func (r *SettingRepositoryImpl) BatchCreate(ctx context.Context, settings []*entities.Setting) error {
	// Early return if no settings to insert
	if len(settings) == 0 {
		return nil
	}

	// Begin a transaction for atomic batch insertion
	// This ensures all settings are inserted together or none at all
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback if not committed (deferred cleanup)
	defer tx.Rollback(ctx)

	// Prepare the batch insert statement
	// Fields: id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at
	query := `INSERT INTO settings (id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Create a batch for efficient insertion
	// pgx.Batch allows multiple queries to be sent in a single network round-trip
	batch := &pgx.Batch{}

	// Queue each setting for batch insertion
	for _, setting := range settings {
		// Convert ModelID UUID to string for database storage
		// Empty string is used when ModelID is nil
		modelIDStr := ""
		if setting.ModelID != nil {
			modelIDStr = setting.ModelID.String()
		}

		// Add this setting to the batch queue
		batch.Queue(query,
			setting.ID.String(), setting.ModelType, modelIDStr, setting.Key, setting.Value,
			setting.CreatedBy, setting.UpdatedBy, setting.CreatedAt, setting.UpdatedAt)
	}

	// Execute the batch using the transaction
	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Check for any errors during batch execution
	// Process each queued query and check for individual errors
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	// Commit the transaction if all operations succeeded
	return tx.Commit(ctx)
}

// GetByID retrieves a setting by its unique identifier.
// Returns nil and error if the setting is not found or has been deleted.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - id: UUID of the setting to retrieve
//
// Returns:
//   - *entities.Setting: Pointer to the found setting entity, or nil if not found
//   - error: Database operation error if any
//
// Implementation details:
// - Filters out deleted settings (deleted_at IS NULL)
// - Handles polymorphic relationship conversion from string back to UUID
// - Uses QueryRow for single result retrieval
func (r *SettingRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error) {
	// SQL query to select a setting by ID, excluding deleted records
	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE id = $1 AND deleted_at IS NULL`

	var setting entities.Setting
	var modelIDStr *string // Use pointer to handle NULL values from database

	// Execute query and scan results into variables
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
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: String key of the setting to retrieve
//
// Returns:
//   - *entities.Setting: Pointer to the found setting entity, or nil if not found
//   - error: Database operation error if any
//
// Implementation details:
// - Filters out deleted settings (deleted_at IS NULL)
// - Handles polymorphic relationship conversion from string back to UUID
// - Uses QueryRow for single result retrieval
func (r *SettingRepositoryImpl) GetByKey(ctx context.Context, key string) (*entities.Setting, error) {
	// SQL query to select a setting by key, excluding deleted records
	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE key = $1 AND deleted_at IS NULL`

	var setting entities.Setting
	var modelIDStr *string // Use pointer to handle NULL values from database

	// Execute query and scan results into variables
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
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - []*entities.Setting: Slice of all non-deleted setting entities
//   - error: Database operation error if any
//
// Implementation details:
// - Filters out deleted settings (deleted_at IS NULL)
// - Orders results by key for consistent output
// - Uses helper method scanSettingRow for row processing
// - Handles multiple rows efficiently
func (r *SettingRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Setting, error) {
	// SQL query to select all non-deleted settings, ordered by key
	query := `SELECT id, model_type, model_id, key, value, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM settings WHERE deleted_at IS NULL
			  ORDER BY key ASC`

	// Execute query to get all matching rows
	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure rows are closed after processing

	// Process each row and build the result slice
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
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - setting: Pointer to the setting entity with updated values
//
// Returns:
//   - error: Database operation error if any
//
// Implementation details:
// - Updates only non-deleted settings (deleted_at IS NULL)
// - Handles ModelID conversion for polymorphic relationships
// - Updates timestamp and user tracking fields
func (r *SettingRepositoryImpl) Update(ctx context.Context, setting *entities.Setting) error {
	// SQL query to update an existing setting, excluding deleted records
	query := `UPDATE settings SET model_type = $1, model_id = $2, key = $3, value = $4, updated_at = $5, updated_by = $6
			  WHERE id = $7 AND deleted_at IS NULL`

	// Convert ModelID UUID to string for database storage
	// Empty string is used when ModelID is nil
	modelIDStr := ""
	if setting.ModelID != nil {
		modelIDStr = setting.ModelID.String()
	}

	// Execute the UPDATE query with all updated fields
	_, err := r.pgxPool.Exec(ctx, query, setting.ModelType, modelIDStr, setting.Key, setting.Value,
		setting.UpdatedAt, setting.UpdatedBy, setting.ID.String())
	return err
}

// Delete performs a soft delete by setting the deleted_at timestamp.
// The record remains in the database but is excluded from normal queries.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - id: UUID of the setting to delete
//
// Returns:
//   - error: Database operation error if any
//
// Implementation details:
// - Uses soft delete approach (sets deleted_at timestamp)
// - Only affects non-deleted records
// - Uses NOW() for current timestamp
func (r *SettingRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// SQL query to soft delete a setting by setting deleted_at timestamp
	query := `UPDATE settings SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

// ExistsByKey checks if a setting with the given key exists and is not deleted.
// Returns true if the setting exists, false otherwise.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: String key to check for existence
//
// Returns:
//   - bool: true if the setting exists and is not deleted, false otherwise
//   - error: Database operation error if any
//
// Implementation details:
// - Uses EXISTS subquery for efficient existence checking
// - Filters out deleted settings
// - Returns boolean result directly
func (r *SettingRepositoryImpl) ExistsByKey(ctx context.Context, key string) (bool, error) {
	// SQL query to check existence using EXISTS subquery
	query := `SELECT EXISTS(SELECT 1 FROM settings WHERE key = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, key).Scan(&exists)
	return exists, err
}

// Count returns the total number of non-deleted settings in the database.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - int64: Total count of non-deleted settings
//   - error: Database operation error if any
//
// Implementation details:
// - Uses COUNT(*) for efficient counting
// - Filters out deleted settings
// - Returns int64 for large dataset compatibility
func (r *SettingRepositoryImpl) Count(ctx context.Context) (int64, error) {
	// SQL query to count all non-deleted settings
	query := `SELECT COUNT(*) FROM settings WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// scanSettingRow is a helper function to scan a setting row from database result set.
// It handles the conversion of model_id from string back to UUID for polymorphic relationships.
// This method is used by GetAll to process multiple rows efficiently.
//
// Parameters:
//   - rows: pgx.Rows containing the database result set
//
// Returns:
//   - *entities.Setting: Pointer to the scanned setting entity
//   - error: Scanning error if any
//
// Implementation details:
// - Handles NULL values for model_id field
// - Converts string model_id back to UUID when present
// - Reuses scanning logic across multiple methods
func (r *SettingRepositoryImpl) scanSettingRow(rows pgx.Rows) (*entities.Setting, error) {
	var setting entities.Setting
	var modelIDStr *string // Use pointer to handle NULL values from database

	// Scan the current row into variables
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
