package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// TagRepository defines the interface for managing tag entities and their relationships
// This repository handles CRUD operations for tags and manages the polymorphic
// many-to-many relationship between tags and other entities (taggables).
type TagRepository interface {
	// Create adds a new tag to the system
	Create(ctx context.Context, tag *entities.Tag) error

	// GetByID retrieves a specific tag by its UUID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error)

	// GetAll retrieves all non-deleted tags from the system
	GetAll(ctx context.Context) ([]*entities.Tag, error)

	// Update modifies an existing tag's information
	Update(ctx context.Context, tag *entities.Tag) error

	// Delete performs a soft delete by setting deleted_at timestamp
	Delete(ctx context.Context, id uuid.UUID) error

	// AttachTag creates a relationship between a tag and a taggable entity
	AttachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error

	// DetachTag removes the relationship between a tag and a taggable entity
	DetachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error

	// GetTagsForEntity retrieves all tags associated with a specific entity
	GetTagsForEntity(ctx context.Context, taggableID uuid.UUID, taggableType string) ([]*entities.Tag, error)
}

// TagRepositoryImpl implements the TagRepository interface
// This struct provides concrete implementations for tag management operations
// using PostgreSQL for persistence and Redis for caching (if needed).
type TagRepositoryImpl struct {
	pgxPool     *pgxpool.Pool // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable // Redis client for caching operations
}

// NewTagRepository creates a new instance of TagRepositoryImpl
// This constructor function initializes the repository with the required dependencies.
//
// Parameters:
//   - pgxPool: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - TagRepository: interface implementation for tag management
func NewTagRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) TagRepository {
	return &TagRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

// Create adds a new tag to the tags table
// This method inserts a new tag record with all required fields including
// generated UUID, timestamps, and user tracking information.
//
// Parameters:
//   - ctx: context for the database operation
//   - tag: pointer to the tag entity to create
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) Create(ctx context.Context, tag *entities.Tag) error {
	// Insert new tag with all required fields
	query := `INSERT INTO tags (id, name, slug, color, created_by, updated_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pgxPool.Exec(ctx, query, tag.ID.String(), tag.Name, tag.Slug, tag.Color, tag.CreatedBy, tag.UpdatedBy, tag.CreatedAt, tag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// GetByID retrieves a specific tag by its UUID from the database
// This method performs a soft-delete aware query, only returning tags that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the tag to retrieve
//
// Returns:
//   - *entities.Tag: pointer to the found tag entity, or nil if not found
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	// Query for tag by ID, excluding soft-deleted tags
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE id = $1 AND deleted_at IS NULL`

	var tag entities.Tag
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return &tag, nil
}

// GetAll retrieves all non-deleted tags from the database
// This method returns a slice of all active tags, ordered by creation date.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - []*entities.Tag: slice of all non-deleted tag entities
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Tag, error) {
	// Query for all non-deleted tags
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE deleted_at IS NULL`
	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	// Iterate through results and build tag entities
	var tags []*entities.Tag
	for rows.Next() {
		tag, err := r.scanTagRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// Update modifies an existing tag's information in the database
// This method updates the tag's name, slug, color, and tracking information.
// Only non-deleted tags can be updated.
//
// Parameters:
//   - ctx: context for the database operation
//   - tag: pointer to the tag entity with updated information
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) Update(ctx context.Context, tag *entities.Tag) error {
	// Update tag fields, excluding soft-deleted tags
	query := `UPDATE tags SET name = $1, slug = $2, color = $3, updated_at = $4, updated_by = $5 WHERE id = $6 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, tag.Name, tag.Slug, tag.Color, tag.UpdatedAt, tag.UpdatedBy, tag.ID.String())
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}
	return nil
}

// Delete performs a soft delete by setting the deleted_at timestamp
// This method doesn't physically remove the record but marks it as deleted
// for data integrity and potential recovery purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the tag to soft delete
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete by setting deleted_at timestamp
	query := `UPDATE tags SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	return nil
}

// AttachTag creates a relationship between a tag and a taggable entity
// This method inserts a record into the taggables table to establish
// the polymorphic many-to-many relationship.
//
// Parameters:
//   - ctx: context for the database operation
//   - tagID: UUID of the tag to attach
//   - taggableID: UUID of the entity to tag
//   - taggableType: string identifier for the type of entity being tagged
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) AttachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error {
	// Create taggable relationship with new UUID and timestamps
	query := `INSERT INTO taggables (id, tag_id, taggable_id, taggable_type, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pgxPool.Exec(ctx, query, uuid.New().String(), tagID.String(), taggableID.String(), taggableType, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to attach tag: %w", err)
	}
	return nil
}

// DetachTag removes the relationship between a tag and a taggable entity
// This method deletes the corresponding record from the taggables table.
//
// Parameters:
//   - ctx: context for the database operation
//   - tagID: UUID of the tag to detach
//   - taggableID: UUID of the entity to untag
//   - taggableType: string identifier for the type of entity being untagged
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) DetachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error {
	// Remove taggable relationship by deleting the record
	query := `DELETE FROM taggables WHERE tag_id = $1 AND taggable_id = $2 AND taggable_type = $3`
	_, err := r.pgxPool.Exec(ctx, query, tagID.String(), taggableID.String(), taggableType)
	if err != nil {
		return fmt.Errorf("failed to detach tag: %w", err)
	}
	return nil
}

// GetTagsForEntity retrieves all tags associated with a specific entity
// This method joins the tags and taggables tables to find all tags
// that are attached to a particular entity of a specific type.
//
// Parameters:
//   - ctx: context for the database operation
//   - taggableID: UUID of the entity to get tags for
//   - taggableType: string identifier for the type of entity
//
// Returns:
//   - []*entities.Tag: slice of tag entities associated with the entity
//   - error: nil if successful, or wrapped error if the operation fails
func (r *TagRepositoryImpl) GetTagsForEntity(ctx context.Context, taggableID uuid.UUID, taggableType string) ([]*entities.Tag, error) {
	// Join tags and taggables tables to find entity tags
	// Filter by taggable entity and exclude soft-deleted tags
	query := `SELECT t.id, t.name, t.slug, t.color, t.created_by, t.updated_by, t.created_at, t.updated_at, t.deleted_at FROM tags t JOIN taggables tg ON t.id = tg.tag_id WHERE tg.taggable_id = $1 AND tg.taggable_type = $2 AND t.deleted_at IS NULL`
	rows, err := r.pgxPool.Query(ctx, query, taggableID.String(), taggableType)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for entity: %w", err)
	}
	defer rows.Close()

	// Iterate through results and build tag entities
	var tags []*entities.Tag
	for rows.Next() {
		tag, err := r.scanTagRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// scanTagRow is a helper function to scan a tag row from database result set
// This method extracts tag data from a database row and constructs a Tag entity.
// It's used by methods that return multiple tags to avoid code duplication.
//
// Parameters:
//   - rows: pgx.Rows containing the database result set
//
// Returns:
//   - *entities.Tag: pointer to the scanned tag entity
//   - error: nil if successful, or error if scanning fails
func (r *TagRepositoryImpl) scanTagRow(rows pgx.Rows) (*entities.Tag, error) {
	var tag entities.Tag
	err := rows.Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.Color)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}
