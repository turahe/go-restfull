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

// PostgresTagRepository implements the TagRepository interface using PostgreSQL as the primary
// data store and Redis for caching. This repository provides CRUD operations for tag entities
// with support for soft deletes, search functionality, and pagination. It embeds the
// BaseTransactionalRepository to inherit transaction management capabilities.
type PostgresTagRepository struct {
	// BaseTransactionalRepository provides transaction management functionality
	*BaseTransactionalRepository
	// db holds the PostgreSQL connection pool for database operations
	db *pgxpool.Pool
	// redisClient holds the Redis client for caching operations
	redisClient redis.Cmdable
}

// NewPostgresTagRepository creates a new PostgreSQL tag repository instance.
// This factory function initializes the repository with database and Redis connections,
// and sets up the base transactional repository for transaction management.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations
//   - redisClient: The Redis client for caching operations
//
// Returns:
//   - repositories.TagRepository: A new tag repository instance
func NewPostgresTagRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.TagRepository {
	return &PostgresTagRepository{BaseTransactionalRepository: NewBaseTransactionalRepository(db), db: db, redisClient: redisClient}
}

// Create persists a new tag entity to the database.
// This method inserts a new tag record with all required fields including
// audit information (created_by, updated_by, timestamps).
//
// Parameters:
//   - ctx: Context for the database operation
//   - tag: The tag entity to create
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) Create(ctx context.Context, tag *entities.Tag) error {
	query := `INSERT INTO tags (id, name, slug, color, created_by, updated_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.Exec(ctx, query, tag.ID, tag.Name, tag.Slug, tag.Color, tag.CreatedBy, tag.UpdatedBy, tag.CreatedAt, tag.UpdatedAt)
	return err
}

// GetByID retrieves a tag entity by its unique identifier.
// This method performs a soft-delete aware query, excluding records that have been
// marked as deleted. It returns the complete tag entity with all fields populated.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the tag to retrieve
//
// Returns:
//   - *entities.Tag: The found tag entity, or nil if not found
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE id = $1 AND deleted_at IS NULL`
	var tag entities.Tag
	if err := r.db.QueryRow(ctx, query, id).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetBySlug retrieves a tag entity by its slug identifier.
// This method is useful for URL-friendly tag lookups and performs a soft-delete
// aware query. The slug is typically a URL-safe version of the tag name.
//
// Parameters:
//   - ctx: Context for the database operation
//   - slug: The slug identifier of the tag to retrieve
//
// Returns:
//   - *entities.Tag: The found tag entity, or nil if not found
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) GetBySlug(ctx context.Context, slug string) (*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE slug = $1 AND deleted_at IS NULL`
	var tag entities.Tag
	if err := r.db.QueryRow(ctx, query, slug).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetAll retrieves a paginated list of all active tags.
// This method returns tags ordered by creation date (newest first) and supports
// pagination through limit and offset parameters. It excludes soft-deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - limit: Maximum number of tags to return
//   - offset: Number of tags to skip for pagination
//
// Returns:
//   - []*entities.Tag: List of tag entities
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []*entities.Tag
	for rows.Next() {
		var tag entities.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

// Search performs a case-insensitive search for tags by name or slug.
// This method uses ILIKE for pattern matching and supports partial matches.
// Results are ordered by creation date (newest first) and support pagination.
// The search is soft-delete aware, excluding deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - query: The search term to match against tag names and slugs
//   - limit: Maximum number of tags to return
//   - offset: Number of tags to skip for pagination
//
// Returns:
//   - []*entities.Tag: List of matching tag entities
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	q := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE deleted_at IS NULL AND (name ILIKE $1 OR slug ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []*entities.Tag
	for rows.Next() {
		var tag entities.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

// Update modifies an existing tag entity in the database.
// This method updates the tag's name, slug, color, and audit fields while
// preserving the original ID and creation information. It only updates
// non-deleted tags.
//
// Parameters:
//   - ctx: Context for the database operation
//   - tag: The tag entity with updated values
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) Update(ctx context.Context, tag *entities.Tag) error {
	query := `UPDATE tags SET name=$1, slug=$2, color=$3, updated_at=$4, updated_by=$5 WHERE id=$6 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, tag.Name, tag.Slug, tag.Color, tag.UpdatedAt, tag.UpdatedBy, tag.ID)
	return err
}

// Delete performs a soft delete of a tag entity.
// This method marks the tag as deleted by setting the deleted_at timestamp
// rather than physically removing the record. This preserves data integrity
// and allows for potential recovery.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the tag to delete
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tags SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ExistsBySlug checks if a tag with the specified slug already exists.
// This method is useful for validation purposes, ensuring tag slugs are unique
// before creation. It performs a soft-delete aware check.
//
// Parameters:
//   - ctx: Context for the database operation
//   - slug: The slug to check for existence
//
// Returns:
//   - bool: True if a tag with the slug exists, false otherwise
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tags WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// Count returns the total number of active tags in the system.
// This method excludes soft-deleted records and is useful for pagination
// calculations and system statistics.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - int64: The total count of active tags
//   - error: Any error that occurred during the database operation
func (r *PostgresTagRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM tags WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
