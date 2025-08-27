// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"context"
	"strings"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresContentRepository implements the ContentRepository interface using PostgreSQL as the primary
// data store and Redis for caching. This repository provides CRUD operations for content entities
// with support for polymorphic relationships, soft deletes, content restoration, and search functionality.
// It embeds the BaseTransactionalRepository to inherit transaction management capabilities.
type PostgresContentRepository struct {
	// BaseTransactionalRepository provides transaction management functionality
	*BaseTransactionalRepository
	// db holds the PostgreSQL connection pool for database operations
	db *pgxpool.Pool
	// redisClient holds the Redis client for caching operations
	redisClient redis.Cmdable
}

// NewPostgresContentRepository creates a new PostgreSQL content repository instance.
// This factory function initializes the repository with database and Redis connections,
// and sets up the base transactional repository for transaction management.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations
//   - redisClient: The Redis client for caching operations
//
// Returns:
//   - repositories.ContentRepository: A new content repository instance
func NewPostgresContentRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.ContentRepository {
	return &PostgresContentRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		redisClient:                 redisClient,
	}
}

// Create persists a new content entity to the database.
// This method inserts a new content record with all required fields including
// audit information (created_by, updated_by, timestamps). Content can be associated
// with different model types through polymorphic relationships.
//
// Parameters:
//   - ctx: Context for the database operation
//   - content: The content entity to create
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) Create(ctx context.Context, content *entities.Content) error {
	query := `INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.Exec(ctx, query,
		content.ID, content.ModelType, content.ModelID, content.ContentRaw, content.ContentHTML,
		content.CreatedBy, content.UpdatedBy, content.CreatedAt, content.UpdatedAt,
	)
	return err
}

// GetByID retrieves a content entity by its unique identifier.
// This method returns the complete content entity with all fields populated,
// including deletion information if the content has been soft-deleted.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the content to retrieve
//
// Returns:
//   - *entities.Content: The found content entity, or nil if not found
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents WHERE id = $1`
	var c entities.Content
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML,
		&c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetByModelTypeAndID retrieves all content entities associated with a specific model.
// This method is useful for polymorphic relationships where content can belong to
// different entity types (e.g., posts, pages, comments). Results are ordered by
// creation date (oldest first) to maintain content sequence.
//
// Parameters:
//   - ctx: Context for the database operation
//   - modelType: The type of model the content belongs to (e.g., "Post", "Page")
//   - modelID: The unique identifier of the specific model instance
//
// Returns:
//   - []*entities.Content: List of content entities associated with the model
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents WHERE model_type = $1 AND model_id = $2 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, modelType, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Content
	for rows.Next() {
		var c entities.Content
		if err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML, &c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}

// GetAll retrieves a paginated list of all content entities.
// This method returns content ordered by creation date (newest first) and supports
// pagination through limit and offset parameters. It includes both active and
// soft-deleted content for administrative purposes.
//
// Parameters:
//   - ctx: Context for the database operation
//   - limit: Maximum number of content entities to return
//   - offset: Number of content entities to skip for pagination
//
// Returns:
//   - []*entities.Content: List of content entities
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Content
	for rows.Next() {
		var c entities.Content
		if err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML, &c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}

// Update modifies an existing content entity in the database.
// This method updates the content's model association, raw content, HTML content,
// and audit fields while preserving the original ID and creation information.
//
// Parameters:
//   - ctx: Context for the database operation
//   - content: The content entity with updated values
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) Update(ctx context.Context, content *entities.Content) error {
	query := `UPDATE contents SET model_type=$1, model_id=$2, content_raw=$3, content_html=$4, updated_by=$5, updated_at=$6 WHERE id=$7`
	_, err := r.db.Exec(ctx, query, content.ModelType, content.ModelID, content.ContentRaw, content.ContentHTML, content.UpdatedBy, content.UpdatedAt, content.ID)
	return err
}

// Delete performs a soft delete of a content entity.
// This method marks the content as deleted by setting the deleted_at timestamp
// and recording who performed the deletion. This preserves data integrity
// and allows for potential recovery while maintaining audit trails.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the content to delete
//   - deletedBy: The user ID who performed the deletion
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	query := `UPDATE contents SET deleted_by=$1, deleted_at=NOW(), updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, deletedBy, id)
	return err
}

// HardDelete permanently removes a content entity from the database.
// This method performs a physical deletion and should be used with caution
// as it cannot be undone. It's typically used for cleanup operations or
// when content must be completely removed for legal/compliance reasons.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the content to permanently delete
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM contents WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Restore recovers a previously soft-deleted content entity.
// This method clears the deletion markers and allows the content to be
// accessed normally again. It updates the audit trail to record who
// performed the restoration.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the content to restore
//   - updatedBy: The user ID who performed the restoration
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) Restore(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	query := `UPDATE contents SET deleted_by=NULL, deleted_at=NULL, updated_by=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, query, updatedBy, id)
	return err
}

// Count returns the total number of active content entities in the system.
// This method excludes soft-deleted records and is useful for pagination
// calculations and system statistics.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - int64: The total count of active content entities
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM contents WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CountByModelType returns the total number of active content entities for a specific model type.
// This method is useful for understanding content distribution across different entity types
// and for pagination calculations within specific model contexts.
//
// Parameters:
//   - ctx: Context for the database operation
//   - modelType: The type of model to count content for (e.g., "Post", "Page")
//
// Returns:
//   - int64: The total count of active content entities for the specified model type
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) CountByModelType(ctx context.Context, modelType string) (int64, error) {
	query := `SELECT COUNT(*) FROM contents WHERE model_type = $1 AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query, modelType).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Search performs a case-insensitive search for content by raw content or HTML content.
// This method uses ILIKE for pattern matching and supports partial matches across
// both content fields. Results are ordered by creation date (newest first) and support
// pagination. The search is soft-delete aware, excluding deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - query: The search term to match against content_raw and content_html fields
//   - limit: Maximum number of matching content entities to return
//   - offset: Number of matching content entities to skip for pagination
//
// Returns:
//   - []*entities.Content: List of matching content entities
//   - error: Any error that occurred during the database operation
func (r *PostgresContentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	q := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents WHERE deleted_at IS NULL AND (content_raw ILIKE $1 OR content_html ILIKE $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Content
	for rows.Next() {
		var c entities.Content
		if err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML, &c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}
