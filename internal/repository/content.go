package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ContentRepository defines the interface for managing content entities
// This repository handles CRUD operations for content items that can be associated
// with various model types (polymorphic relationship) such as posts, pages, etc.
type ContentRepository interface {
	// Create adds a new content item to the system
	Create(ctx context.Context, content *entities.Content) error

	// GetByID retrieves a specific content item by its UUID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error)

	// GetAll retrieves content items with pagination support
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error)

	// GetByModelTypeAndID retrieves all content items for a specific model entity
	GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error)

	// Search performs full-text search on content using ILIKE pattern matching
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error)

	// Update modifies an existing content item's information
	Update(ctx context.Context, content *entities.Content) error

	// Delete permanently removes a content item from the system
	Delete(ctx context.Context, id uuid.UUID) error

	// Count returns the total number of content items in the system
	Count(ctx context.Context) (int64, error)

	// CountByModelTypeAndID returns the count of content items for a specific model entity
	CountByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) (int64, error)
}

// ContentRepositoryImpl implements the ContentRepository interface
// This struct provides concrete implementations for content management operations
// using PostgreSQL for persistence and Redis for caching (if needed).
type ContentRepositoryImpl struct {
	pgxPool     *pgxpool.Pool // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable // Redis client for caching operations
}

// NewContentRepository creates a new instance of ContentRepositoryImpl
// This constructor function initializes the repository with the required dependencies.
//
// Parameters:
//   - pgxPool: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - ContentRepository: interface implementation for content management
func NewContentRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) ContentRepository {
	return &ContentRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

// Create adds a new content item to the contents table
// This method inserts a new content record with all required fields including
// the polymorphic relationship to other entities.
//
// Parameters:
//   - ctx: context for the database operation
//   - content: pointer to the content entity to create
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) Create(ctx context.Context, content *entities.Content) error {
	// Insert new content with polymorphic relationship fields
	query := `INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pgxPool.Exec(ctx, query,
		content.ID.String(), content.ModelType, content.ModelID.String(), content.ContentRaw, content.ContentHTML,
		content.CreatedBy, content.UpdatedBy, content.CreatedAt, content.UpdatedAt)
	return err
}

// GetByID retrieves a specific content item by its UUID from the database
// This method returns the complete content information including raw and HTML versions.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the content item to retrieve
//
// Returns:
//   - *entities.Content: pointer to the found content entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	// Query for content by ID with all fields
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at
			  FROM contents WHERE id = $1`

	var content entities.Content
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML,
		&content.CreatedBy, &content.UpdatedBy, &content.CreatedAt, &content.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

// GetAll retrieves content items with pagination support
// This method returns content items ordered by creation date (newest first)
// and supports limit/offset for efficient pagination.
//
// Parameters:
//   - ctx: context for the database operation
//   - limit: maximum number of content items to return
//   - offset: number of content items to skip (for pagination)
//
// Returns:
//   - []*entities.Content: slice of content entities with pagination
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
	// Query for content with pagination, ordered by creation date descending
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at
			  FROM contents
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build content entities
	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContentRow(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// GetByModelTypeAndID retrieves all content items for a specific model entity
// This method supports the polymorphic relationship by finding content items
// associated with a particular model type and ID (e.g., all content for a specific post).
//
// Parameters:
//   - ctx: context for the database operation
//   - modelType: string identifier for the type of model (e.g., "post", "page")
//   - modelID: UUID of the specific model entity
//
// Returns:
//   - []*entities.Content: slice of content entities for the specified model
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	// Query for content by model type and ID, ordered by creation date ascending
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at
			  FROM contents WHERE model_type = $1 AND model_id = $2
			  ORDER BY created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, modelType, modelID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build content entities
	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContentRow(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// Search performs full-text search on content using ILIKE pattern matching
// This method searches both raw content and HTML content fields for matches,
// supporting case-insensitive pattern matching with wildcards.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search term to look for in content
//   - limit: maximum number of results to return
//   - offset: number of results to skip (for pagination)
//
// Returns:
//   - []*entities.Content: slice of content entities matching the search query
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	// Search query using ILIKE for case-insensitive pattern matching
	// Search both raw content and HTML content fields
	searchQuery := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at
			  FROM contents 
			  WHERE (content_raw ILIKE $1 OR content_html ILIKE $1)
			  ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	// Add wildcards for pattern matching
	searchTerm := "%" + query + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and build content entities
	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContentRow(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// Update modifies an existing content item's information in the database
// This method updates all content fields including the polymorphic relationship
// and tracking information.
//
// Parameters:
//   - ctx: context for the database operation
//   - content: pointer to the content entity with updated information
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) Update(ctx context.Context, content *entities.Content) error {
	// Update content fields including polymorphic relationship
	query := `UPDATE contents SET model_type = $1, model_id = $2, content_raw = $3, content_html = $4, updated_at = $5, updated_by = $6
			  WHERE id = $7`

	_, err := r.pgxPool.Exec(ctx, query, content.ModelType, content.ModelID.String(), content.ContentRaw, content.ContentHTML,
		content.UpdatedAt, content.UpdatedBy, content.ID.String())
	return err
}

// Delete permanently removes a content item from the database
// This method performs a hard delete, completely removing the content record.
// Use with caution as this operation cannot be undone.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the content item to delete
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// Hard delete the content record
	query := `DELETE FROM contents WHERE id = $1`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

// Count returns the total number of content items in the system
// This method provides a quick count for pagination and reporting purposes.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - int64: total count of content items in the system
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) Count(ctx context.Context) (int64, error) {
	// Use COUNT(*) for efficient counting
	query := `SELECT COUNT(*) FROM contents`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountByModelTypeAndID returns the count of content items for a specific model entity
// This method supports the polymorphic relationship by counting content items
// associated with a particular model type and ID.
//
// Parameters:
//   - ctx: context for the database operation
//   - modelType: string identifier for the type of model (e.g., "post", "page")
//   - modelID: UUID of the specific model entity
//
// Returns:
//   - int64: count of content items for the specified model
//   - error: nil if successful, or database error if the operation fails
func (r *ContentRepositoryImpl) CountByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) (int64, error) {
	// Count content items by model type and ID
	query := `SELECT COUNT(*) FROM contents WHERE model_type = $1 AND model_id = $2`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query, modelType, modelID.String()).Scan(&count)
	return count, err
}

// scanContentRow is a helper function to scan a content row from database result set
// This method extracts content data from a database row and constructs a Content entity.
// It's used by methods that return multiple content items to avoid code duplication.
//
// Parameters:
//   - rows: pgx.Rows containing the database result set
//
// Returns:
//   - *entities.Content: pointer to the scanned content entity
//   - error: nil if successful, or error if scanning fails
func (r *ContentRepositoryImpl) scanContentRow(rows pgx.Rows) (*entities.Content, error) {
	var content entities.Content
	err := rows.Scan(
		&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML,
		&content.CreatedBy, &content.UpdatedBy, &content.CreatedAt, &content.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &content, nil
}
