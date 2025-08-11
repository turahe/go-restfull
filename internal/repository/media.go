package repository

import (
	"context"
	"fmt"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// MediaRepository defines the interface for managing media entities
// This repository handles CRUD operations for media files with support for
// nested set tree structure, user ownership, and soft deletes.
type MediaRepository interface {
	// Create adds a new media item to the system with nested set positioning
	Create(ctx context.Context, media *entities.Media) error

	// GetByID retrieves a specific media item by its UUID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error)

	// GetAll retrieves all non-deleted media items with pagination support
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error)

	// GetByUserID retrieves media items owned by a specific user with pagination
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error)

	// Update modifies an existing media item's information
	Update(ctx context.Context, media *entities.Media) error

	// Delete performs a soft delete by setting deleted_at timestamp
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByFilename checks if a media item with the given filename exists
	ExistsByFilename(ctx context.Context, filename string) (bool, error)

	// Count returns the total number of non-deleted media items
	Count(ctx context.Context) (int64, error)

	// CountByUserID returns the count of media items owned by a specific user
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// Search performs full-text search on media filenames, names, and MIME types
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error)
}

// mediaRepository implements the MediaRepository interface
// This struct provides concrete implementations for media management operations
// using PostgreSQL for persistence, Redis for caching, and nested set for tree structure.
type mediaRepository struct {
	db          *pgxpool.Pool               // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable               // Redis client for caching operations
	nestedSet   *nestedset.NestedSetManager // Manager for nested set tree operations
}

// NewMediaRepository creates a new instance of mediaRepository
// This constructor function initializes the repository with the required dependencies
// including the nested set manager for tree structure operations.
//
// Parameters:
//   - pgxPool: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - MediaRepository: interface implementation for media management
func NewMediaRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) MediaRepository {
	return &mediaRepository{
		db:          pgxPool,
		redisClient: redisClient,
		nestedSet:   nestedset.NewNestedSetManager(pgxPool),
	}
}

// Create adds a new media item to the media table with nested set positioning
// This method calculates the appropriate tree position using nested set values
// and inserts the media record with all required fields including tree structure.
//
// Parameters:
//   - ctx: context for the database operation
//   - media: pointer to the media entity to create
//
// Returns:
//   - error: nil if successful, or wrapped error if the operation fails
func (r *mediaRepository) Create(ctx context.Context, media *entities.Media) error {
	// Calculate nested set values for tree positioning
	// For media, we'll treat it as a flat structure initially
	// If we need to implement folder-like hierarchy later, we can add parent_id support
	values, err := r.nestedSet.CreateNode(ctx, "media", nil, 1)
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign computed nested set values to the entity
	media.RecordLeft = &values.Left
	media.RecordRight = &values.Right
	media.RecordDepth = &values.Depth
	media.RecordOrdering = &values.Ordering

	// Insert the new media with all fields including nested set values
	query := `
		INSERT INTO media (
			id, name, file_name, hash, disk, mime_type, size,
			record_left, record_right, record_depth, record_ordering,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err = r.db.Exec(ctx, query,
		media.ID, media.Name, media.FileName, media.Hash, media.Disk, media.MimeType, media.Size,
		media.RecordLeft, media.RecordRight, media.RecordDepth, media.RecordOrdering,
		media.CreatedBy, media.UpdatedBy, media.CreatedAt, media.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	return nil
}

// GetByID retrieves a specific media item by its UUID from the database
// This method performs a soft-delete aware query, only returning media that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the media item to retrieve
//
// Returns:
//   - *entities.Media: pointer to the found media entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error) {
	// Query for media by ID, excluding soft-deleted media
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE id = $1 AND deleted_at IS NULL
	`

	var media entities.Media
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
		&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
		&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &media, nil
}

// GetAll retrieves all non-deleted media items from the database with pagination support
// This method returns media items ordered by nested set left value for tree traversal
// and supports limit/offset for efficient pagination.
//
// Parameters:
//   - ctx: context for the database operation
//   - limit: maximum number of media items to return
//   - offset: number of media items to skip (for pagination)
//
// Returns:
//   - []*entities.Media: slice of media entities with pagination
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error) {
	// Query for all non-deleted media with pagination, ordered by nested set left value
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE deleted_at IS NULL
		ORDER BY record_left ASC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build media entities
	var mediaList []*entities.Media
	for rows.Next() {
		media, err := r.scanMediaRow(rows)
		if err != nil {
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

// GetByUserID retrieves media items owned by a specific user with pagination support
// This method filters media by the user who created it and orders by nested set structure.
//
// Parameters:
//   - ctx: context for the database operation
//   - userID: UUID of the user whose media to retrieve
//   - limit: maximum number of media items to return
//   - offset: number of media items to skip (for pagination)
//
// Returns:
//   - []*entities.Media: slice of media entities owned by the user
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error) {
	// Query for media by user ID with pagination, ordered by nested set left value
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build media entities
	var mediaList []*entities.Media
	for rows.Next() {
		media, err := r.scanMediaRow(rows)
		if err != nil {
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

// Update modifies an existing media item's information in the database
// This method updates basic media fields but preserves the tree structure.
// Only non-deleted media can be updated.
//
// Parameters:
//   - ctx: context for the database operation
//   - media: pointer to the media entity with updated information
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) Update(ctx context.Context, media *entities.Media) error {
	// For updates, we only update basic fields, not the tree structure
	// This preserves the nested set positioning while allowing content updates
	query := `
		UPDATE media SET name = $1, file_name = $2, hash = $3, disk = $4, 
		                mime_type = $5, size = $6, updated_by = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		media.Name, media.FileName, media.Hash, media.Disk, media.MimeType, media.Size,
		media.UpdatedBy, media.UpdatedAt, media.ID.String())
	return err
}

// Delete performs a soft delete by setting the deleted_at timestamp
// This method doesn't physically remove the record but marks it as deleted
// for data integrity and potential recovery purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the media item to soft delete
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete by setting deleted_at timestamp
	query := `UPDATE media SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id.String())
	return err
}

// ExistsByFilename checks if a media item with the given filename exists and is not deleted
// This method uses an EXISTS subquery for efficient checking without retrieving full data.
//
// Parameters:
//   - ctx: context for the database operation
//   - filename: string filename to check for existence
//
// Returns:
//   - bool: true if the media exists, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) ExistsByFilename(ctx context.Context, filename string) (bool, error) {
	// Use EXISTS subquery for efficient filename existence checking
	query := `SELECT EXISTS(SELECT 1 FROM media WHERE file_name = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, filename).Scan(&exists)
	return exists, err
}

// Count returns the total number of non-deleted media items in the database
// This method provides a quick count for pagination and reporting purposes.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - int64: total count of non-deleted media items
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) Count(ctx context.Context) (int64, error) {
	// Count all non-deleted media items
	query := `SELECT COUNT(*) FROM media WHERE deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountByUserID returns the count of media items owned by a specific user
// This method provides a quick count for user-specific pagination and reporting.
//
// Parameters:
//   - ctx: context for the database operation
//   - userID: UUID of the user whose media count to retrieve
//
// Returns:
//   - int64: count of media items owned by the specified user
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	// Count media items by user ID, excluding soft-deleted items
	query := `SELECT COUNT(*) FROM media WHERE created_by = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, userID.String()).Scan(&count)
	return count, err
}

// Search performs full-text search on media filenames, names, and MIME types
// This method uses ILIKE pattern matching for case-insensitive search with wildcards
// and orders results by nested set structure for consistent tree traversal.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search term to look for in media fields
//   - limit: maximum number of results to return
//   - offset: number of results to skip (for pagination)
//
// Returns:
//   - []*entities.Media: slice of media entities matching the search query
//   - error: nil if successful, or database error if the operation fails
func (r *mediaRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	// Search query using ILIKE for case-insensitive pattern matching
	// Search across filename, name, and MIME type fields
	searchQuery := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE deleted_at IS NULL 
		AND (file_name ILIKE $1 OR name ILIKE $1 OR mime_type ILIKE $1)
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`

	// Add wildcards for pattern matching
	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build media entities
	var mediaList []*entities.Media
	for rows.Next() {
		media, err := r.scanMediaRow(rows)
		if err != nil {
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

// scanMediaRow is a helper function to scan a media row from database result set
// This method extracts media data from a database row and constructs a Media entity.
// It's used by methods that return multiple media items to avoid code duplication.
//
// Parameters:
//   - rows: pgx.Rows containing the database result set
//
// Returns:
//   - *entities.Media: pointer to the scanned media entity
//   - error: nil if successful, or error if scanning fails
func (r *mediaRepository) scanMediaRow(rows pgx.Rows) (*entities.Media, error) {
	var media entities.Media
	err := rows.Scan(
		&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
		&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
		&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &media, nil
}
