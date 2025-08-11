package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/cache"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// postgresPostRepository implements PostRepository interface
// This struct provides concrete implementations for post management operations
// using PostgreSQL for persistence and Redis for caching operations.
type postgresPostRepository struct {
	db          *pgxpool.Pool // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable // Redis client for caching operations
}

// NewPostgresPostRepository creates a new PostgreSQL post repository
// This constructor function initializes the repository with the required dependencies
// including PostgreSQL connection pool and Redis client for caching.
//
// Parameters:
//   - db: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - repositories.PostRepository: interface implementation for post management
func NewPostgresPostRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.PostRepository {
	return &postgresPostRepository{
		db:          db,
		redisClient: redisClient,
	}
}

// Create creates a new post in the database
// This method inserts a new post record with all required fields and invalidates
// the post cache to ensure data consistency.
//
// Parameters:
//   - ctx: context for the database operation
//   - post: pointer to the post entity to create
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *postgresPostRepository) Create(ctx context.Context, post *entities.Post) error {
	query := `
		INSERT INTO posts (id, title, slug, subtitle, description, type, is_sticky, language, layout, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, query,
		post.ID,
		post.Title,
		post.Slug,
		post.Subtitle,
		post.Description,
		"post", // default type to satisfy NOT NULL constraint
		post.IsSticky,
		post.Language,
		post.Layout,
		post.PublishedAt,
		post.CreatedAt,
		post.UpdatedAt,
	)

	if err == nil {
		// Invalidate post cache
		cache.InvalidatePattern(ctx, cache.PATTERN_POST_CACHE)
	}

	return err
}

// GetByID retrieves a post by ID from the database
// This method first attempts to retrieve the post from cache, and if not found,
// queries the database and caches the result for future requests.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the post to retrieve
//
// Returns:
//   - *entities.Post: pointer to the found post entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *postgresPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf(cache.KEY_POST_BY_ID, id.String())
	var post entities.Post
	err := cache.GetJSON(ctx, cacheKey, &post)
	if err == nil {
		return &post, nil
	}

	query := `
		SELECT id, title, slug, subtitle, description, is_sticky, language, layout, published_at, created_at, updated_at, deleted_at
		FROM posts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var publishedAt sql.NullTime
	var deletedAt sql.NullTime

	err = r.db.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Slug,
		&post.Subtitle,
		&post.Description,
		&post.IsSticky,
		&post.Language,
		&post.Layout,
		&publishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		return nil, err
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}
	if deletedAt.Valid {
		post.DeletedAt = &deletedAt.Time
	}

	// Cache the result
	cache.SetJSON(ctx, cacheKey, &post, cache.DefaultCacheDuration)

	return &post, nil
}

// GetBySlug retrieves a post by slug
func (r *postgresPostRepository) GetBySlug(ctx context.Context, slug string) (*entities.Post, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf(cache.KEY_POST_BY_SLUG, slug)
	var post entities.Post
	err := cache.GetJSON(ctx, cacheKey, &post)
	if err == nil {
		return &post, nil
	}

	query := `
		SELECT id, title, slug, subtitle, description, is_sticky, language, layout, published_at, created_at, updated_at, deleted_at
		FROM posts
		WHERE slug = $1 AND deleted_at IS NULL
	`

	var publishedAt sql.NullTime
	var deletedAt sql.NullTime

	err = r.db.QueryRow(ctx, query, slug).Scan(
		&post.ID,
		&post.Title,
		&post.Slug,
		&post.Subtitle,
		&post.Description,
		&post.IsSticky,
		&post.Language,
		&post.Layout,
		&publishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		return nil, err
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}
	if deletedAt.Valid {
		post.DeletedAt = &deletedAt.Time
	}

	// Cache the result
	cache.SetJSON(ctx, cacheKey, &post, cache.DefaultCacheDuration)

	return &post, nil
}

// GetByAuthor retrieves posts by author ID
func (r *postgresPostRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.Post, error) {
	query := `
		SELECT id, title, slug, subtitle, description, is_sticky, language, layout, published_at, created_at, updated_at, deleted_at
		FROM posts
		WHERE author_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// GetAll retrieves all posts with pagination
func (r *postgresPostRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	query := `
		SELECT id, title, slug, subtitle, description, is_sticky, language, layout, published_at, created_at, updated_at, deleted_at
		FROM posts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// GetPublished retrieves only published posts
func (r *postgresPostRepository) GetPublished(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	query := `
		SELECT id, title, slug, subtitle, description, is_sticky, language, layout, published_at, created_at, updated_at, deleted_at
		FROM posts
		WHERE status = 'published' AND deleted_at IS NULL
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// Search searches posts by query
func (r *postgresPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	searchQuery := `
		SELECT id, title, slug, subtitle, description, is_sticky, language, layout, published_at, created_at, updated_at, deleted_at
		FROM posts
		WHERE (title ILIKE $1 OR slug ILIKE $1 OR subtitle ILIKE $1 OR description ILIKE $1) AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchTerm := fmt.Sprintf("%%%s%%", query)
	rows, err := r.db.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// Update updates an existing post
func (r *postgresPostRepository) Update(ctx context.Context, post *entities.Post) error {
	query := `
        UPDATE posts
        SET title = $1, slug = $2, subtitle = $3, description = $4, is_sticky = $5, language = $6, layout = $7, updated_at = $8
        WHERE id = $9 AND deleted_at IS NULL
    `

	result, err := r.db.Exec(ctx, query,
		post.Title,
		post.Slug,
		post.Subtitle,
		post.Description,
		post.IsSticky,
		post.Language,
		post.Layout,
		post.UpdatedAt,
		post.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}

	// Invalidate post cache
	cache.InvalidatePattern(ctx, cache.PATTERN_POST_CACHE)

	return nil
}

// Delete soft deletes a post
func (r *postgresPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE posts
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}

	// Invalidate post cache
	cache.InvalidatePattern(ctx, cache.PATTERN_POST_CACHE)

	return nil
}

// Publish publishes a post
func (r *postgresPostRepository) Publish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE posts
		SET published_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}

	// Invalidate post cache
	cache.InvalidatePattern(ctx, cache.PATTERN_POST_CACHE)

	return nil
}

// Unpublish unpublishes a post
func (r *postgresPostRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE posts
		SET published_at = NULL, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}

	// Invalidate post cache
	cache.InvalidatePattern(ctx, cache.PATTERN_POST_CACHE)

	return nil
}

// Count returns the total number of posts
func (r *postgresPostRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountPublished returns the total number of published posts
func (r *postgresPostRepository) CountPublished(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM posts WHERE published_at IS NOT NULL AND deleted_at IS NULL`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountBySearch returns the total number of posts matching the search query
func (r *postgresPostRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
	searchQuery := `
		SELECT COUNT(*) FROM posts 
		WHERE deleted_at IS NULL 
		AND (title ILIKE $1 OR slug ILIKE $1)
	`
	var count int64
	err := r.db.QueryRow(ctx, searchQuery, fmt.Sprintf("%%%s%%", query)).Scan(&count)
	return count, err
}

// CountBySearchPublished returns the total number of published posts matching the search query
func (r *postgresPostRepository) CountBySearchPublished(ctx context.Context, query string) (int64, error) {
	searchQuery := `
		SELECT COUNT(*) FROM posts 
		WHERE published_at IS NOT NULL AND deleted_at IS NULL AND (title ILIKE $1 OR slug ILIKE $1)
	`
	var count int64
	err := r.db.QueryRow(ctx, searchQuery, fmt.Sprintf("%%%s%%", query)).Scan(&count)
	return count, err
}

// SearchPublished searches published posts by query
func (r *postgresPostRepository) SearchPublished(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	searchQuery := `
		SELECT id, title, slug, subtitle, description, is_sticky, language, layout, published_at, created_at, updated_at, deleted_at
		FROM posts
		WHERE published_at IS NOT NULL AND deleted_at IS NULL 
		AND (title ILIKE $1 OR slug ILIKE $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, searchQuery, fmt.Sprintf("%%%s%%", query), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*entities.Post
	for rows.Next() {
		var post entities.Post
		var publishedAt sql.NullTime
		var deletedAt sql.NullTime

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Subtitle,
			&post.Description,
			&post.IsSticky,
			&post.Language,
			&post.Layout,
			&publishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		if publishedAt.Valid {
			post.PublishedAt = &publishedAt.Time
		}
		if deletedAt.Valid {
			post.DeletedAt = &deletedAt.Time
		}

		posts = append(posts, &post)
	}

	return posts, nil
}

// BeginTransaction starts a new transaction for this repository
func (r *postgresPostRepository) BeginTransaction(ctx context.Context) (repositories.Transaction, error) {
	// This is a placeholder implementation since the actual transaction management
	// is handled by the adapter layer through BaseTransactionalRepository
	// The concrete repository doesn't need to implement transaction logic
	return nil, fmt.Errorf("transactions should be handled through the adapter layer")
}

// WithTransaction executes repository operations within a transaction
func (r *postgresPostRepository) WithTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	// This is a placeholder implementation since the actual transaction management
	// is handled by the adapter layer through BaseTransactionalRepository
	// The concrete repository doesn't need to implement transaction logic
	return fmt.Errorf("transactions should be handled through the adapter layer")
}

// scanPostFromScanner scans a single post from a row/rows scanner
func (r *postgresPostRepository) scanPostFromScanner(scanner interface{ Scan(dest ...any) error }) (*entities.Post, error) {
	var post entities.Post
	var publishedAt sql.NullTime
	var deletedAt sql.NullTime

	if err := scanner.Scan(
		&post.ID,
		&post.Title,
		&post.Slug,
		&post.Subtitle,
		&post.Description,
		&post.IsSticky,
		&post.Language,
		&post.Layout,
		&publishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}
	if deletedAt.Valid {
		post.DeletedAt = &deletedAt.Time
	}
	return &post, nil
}

// scanPosts scans all rows into a slice of posts
func (r *postgresPostRepository) scanPosts(rows pgx.Rows) ([]*entities.Post, error) {
	var posts []*entities.Post
	for rows.Next() {
		p, err := r.scanPostFromScanner(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}
