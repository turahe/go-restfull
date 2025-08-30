package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/gomarkdown/markdown"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresPostRepository provides the concrete implementation of the PostRepository interface
// using PostgreSQL as the underlying data store. This struct handles all post-related
// database operations including CRUD operations, search, and post management.
type PostgresPostRepository struct {
	*BaseTransactionalRepository
	db *pgxpool.Pool // PostgreSQL connection pool for database operations
}

// NewPostgresPostRepository creates a new instance of PostgresPostRepository
// This constructor function initializes the repository with the required dependencies.
//
// Parameters:
//   - db: PostgreSQL connection pool for database operations
//
// Returns:
//   - repositories.PostRepository: interface implementation for post management
func NewPostgresPostRepository(db *pgxpool.Pool) repositories.PostRepository {
	return &PostgresPostRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
	}
}

// CreateWithContent persists a new post to the database with its associated content
// This method uses a database transaction to ensure both post and content
// are inserted atomically, maintaining data consistency.
// The markdown content is automatically converted to HTML for storage.
//
// Parameters:
//   - ctx: context for the database operation
//   - post: pointer to the post entity to create
//   - contentRaw: raw markdown content text
//   - createdBy: UUID of the user creating the post
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) Create(ctx context.Context, post *entities.Post) error {
	// Convert markdown to HTML
	contentHTML := string(markdown.ToHTML([]byte(post.Content), nil, nil))

	// Use a transaction to ensure both post and content are inserted atomically
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			// Rollback on error
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				// Log rollback error but return the original error
				fmt.Printf("failed to rollback transaction: %v\n", rollbackErr)
			}
		}
	}()

	// Insert post
	postQuery := `
		INSERT INTO posts (
			id, title, slug, subtitle, description, type, language, layout,
			is_sticky, published_at, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err = tx.Exec(ctx, postQuery,
		post.ID,
		post.Title,
		post.Slug,
		post.Subtitle,
		post.Description,
		post.Type,
		post.Language,
		post.Layout,
		post.IsSticky,
		post.PublishedAt,
		post.CreatedAt,
		post.UpdatedAt,
		post.CreatedBy,
		post.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to insert post: %w", err)
	}

	// Insert content
	contentQuery := `
		INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	contentID := uuid.New()
	now := time.Now()
	_, err = tx.Exec(ctx, contentQuery,
		contentID,
		"post", // model type is "post"
		post.ID,
		post.Content,
		contentHTML,
		post.CreatedBy,
		post.UpdatedBy,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert content: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a post by its unique identifier
// This method performs a soft-delete aware query, only returning posts that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the post to retrieve
//
// Returns:
//   - *entities.Post: pointer to the found post entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.type, p.language, p.layout,
			   p.is_sticky, p.published_at, p.created_at, p.updated_at, p.deleted_at,
			   c.content_raw
		FROM posts p
		LEFT JOIN LATERAL (
			SELECT content_raw 
			FROM contents 
			WHERE model_type = 'post' AND model_id = p.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) c ON true
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	var post entities.Post
	var contentRaw *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Slug,
		&post.Subtitle,
		&post.Description,
		&post.Type,
		&post.Language,
		&post.Layout,
		&post.IsSticky,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.DeletedAt,
		&contentRaw,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post by ID: %w", err)
	}

	// Set content from the joined content table
	if contentRaw != nil {
		post.Content = *contentRaw
	}

	return &post, nil
}

// GetBySlug retrieves a post by its slug
// This method performs a soft-delete aware query, only returning posts that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - slug: slug of the post to retrieve
//
// Returns:
//   - *entities.Post: pointer to the found post entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) GetBySlug(ctx context.Context, slug string) (*entities.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.type, p.language, p.layout,
			   p.is_sticky, p.published_at, p.created_at, p.updated_at, p.deleted_at,
			   c.content_raw
		FROM posts p
		LEFT JOIN LATERAL (
			SELECT content_raw 
			FROM contents 
			WHERE model_type = 'post' AND model_id = p.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) c ON true
		WHERE p.slug = $1 AND p.deleted_at IS NULL
	`

	var post entities.Post
	var contentRaw *string
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&post.ID,
		&post.Title,
		&post.Slug,
		&post.Subtitle,
		&post.Description,
		&post.Type,
		&post.Language,
		&post.Layout,
		&post.IsSticky,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.DeletedAt,
		&contentRaw,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post by slug: %w", err)
	}

	// Set content from the joined content table
	if contentRaw != nil {
		post.Content = *contentRaw
	}

	return &post, nil
}

// GetByAuthor retrieves posts by author ID
// This method returns posts ordered by creation date, with pagination support.
//
// Parameters:
//   - ctx: context for the database operation
//   - authorID: UUID of the author
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Post: slice of post entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.type, p.language, p.layout,
			   p.is_sticky, p.published_at, p.created_at, p.updated_at, p.deleted_at,
			   c.content_raw
		FROM posts p
		LEFT JOIN LATERAL (
			SELECT content_raw 
			FROM contents 
			WHERE model_type = 'post' AND model_id = p.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) c ON true
		WHERE p.author_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by author: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// GetAll retrieves all posts with optional pagination
// This method returns posts ordered by creation date.
//
// Parameters:
//   - ctx: context for the database operation
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Post: slice of post entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.type, p.language, p.layout,
			   p.is_sticky, p.published_at, p.created_at, p.updated_at, p.deleted_at,
			   c.content_raw
		FROM posts p
		LEFT JOIN LATERAL (
			SELECT content_raw 
			FROM contents 
			WHERE model_type = 'post' AND model_id = p.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) c ON true
		WHERE p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all posts: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// GetPublished retrieves published posts with pagination
// This method returns only published posts ordered by publication date.
//
// Parameters:
//   - ctx: context for the database operation
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Post: slice of published post entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) GetPublished(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.type, p.language, p.layout,
			   p.is_sticky, p.published_at, p.created_at, p.updated_at, p.deleted_at,
			   c.content_raw
		FROM posts p
		LEFT JOIN LATERAL (
			SELECT content_raw 
			FROM contents 
			WHERE model_type = 'post' AND model_id = p.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) c ON true
		WHERE p.published_at IS NOT NULL AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get published posts: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// Search searches posts by query
// This method performs a case-insensitive search across title, subtitle, and description fields.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search query string
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Post: slice of matching post entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	searchQuery := `
		SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.type, p.language, p.layout,
			   p.is_sticky, p.published_at, p.created_at, p.updated_at, p.deleted_at,
			   c.content_raw
		FROM posts p
		LEFT JOIN LATERAL (
			SELECT content_raw 
			FROM contents 
			WHERE model_type = 'post' AND model_id = p.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) c ON true
		WHERE p.deleted_at IS NULL AND (
			p.title ILIKE $1 OR p.subtitle ILIKE $1 OR p.description ILIKE $1
		)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// Update updates an existing post in the database
// This method updates all post fields and sets the updated_at timestamp.
//
// Parameters:
//   - ctx: context for the database operation
//   - post: pointer to the post entity with updated values
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) Update(ctx context.Context, post *entities.Post) error {
	query := `
		UPDATE posts SET
			title = $2, slug = $3, subtitle = $4, description = $5, type = $6,
			language = $7, layout = $8, is_sticky = $9, published_at = $10,
			updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		post.ID,
		post.Title,
		post.Slug,
		post.Subtitle,
		post.Description,
		post.Type,
		post.Language,
		post.Layout,
		post.IsSticky,
		post.PublishedAt,
		post.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found or already deleted")
	}

	return nil
}

// Delete performs a soft delete of a post by setting the deleted_at timestamp
// This method preserves the data while marking it as deleted for business logic purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the post to delete
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE posts SET
			deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found or already deleted")
	}

	return nil
}

// Publish publishes a post by setting the published_at timestamp
// This method makes the post publicly visible.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the post to publish
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) Publish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE posts SET
			published_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to publish post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found or already deleted")
	}

	return nil
}

// Unpublish unpublishes a post by clearing the published_at timestamp
// This method makes the post no longer publicly visible.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the post to unpublish
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE posts SET
			published_at = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unpublish post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found or already deleted")
	}

	return nil
}

// Count returns the total number of posts
// This method is useful for pagination and reporting purposes.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - int64: total count of posts
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) Count(ctx context.Context) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM posts
		WHERE deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}

	return count, nil
}

// CountPublished returns the total number of published posts
// This method is useful for reporting and analytics purposes.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - int64: total count of published posts
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) CountPublished(ctx context.Context) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM posts
		WHERE published_at IS NOT NULL AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count published posts: %w", err)
	}

	return count, nil
}

// CountBySearch returns the total number of posts matching the search query
// This method is useful for pagination when searching posts.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search query string
//
// Returns:
//   - int64: total count of posts matching the search
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
	searchQuery := `
		SELECT COUNT(*)
		FROM posts
		WHERE deleted_at IS NULL AND (
			title ILIKE $1 OR subtitle ILIKE $1 OR description ILIKE $1
		)
	`

	searchPattern := "%" + query + "%"
	var count int64
	err := r.db.QueryRow(ctx, searchQuery, searchPattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts by search: %w", err)
	}

	return count, nil
}

// CountBySearchPublished returns the total number of published posts matching the search query
// This method is useful for pagination when searching published posts.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search query string
//
// Returns:
//   - int64: total count of published posts matching the search
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) CountBySearchPublished(ctx context.Context, query string) (int64, error) {
	searchQuery := `
		SELECT COUNT(*)
		FROM posts
		WHERE published_at IS NOT NULL AND deleted_at IS NULL AND (
			title ILIKE $1 OR subtitle ILIKE $1 OR description ILIKE $1
		)
	`

	searchPattern := "%" + query + "%"
	var count int64
	err := r.db.QueryRow(ctx, searchQuery, searchPattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count published posts by search: %w", err)
	}

	return count, nil
}

// SearchPublished searches published posts by query
// This method performs a case-insensitive search across title, subtitle, and description fields
// for published posts only.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search query string
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Post: slice of matching published post entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) SearchPublished(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	searchQuery := `
		SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.type, p.language, p.layout,
			   p.is_sticky, p.published_at, p.created_at, p.updated_at, p.deleted_at,
			   c.content_raw
		FROM posts p
		LEFT JOIN LATERAL (
			SELECT content_raw 
			FROM contents 
			WHERE model_type = 'post' AND model_id = p.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) c ON true
		WHERE p.published_at IS NOT NULL AND p.deleted_at IS NULL AND (
			p.title ILIKE $1 OR p.subtitle ILIKE $1 OR p.description ILIKE $1
		)
		ORDER BY p.published_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search published posts: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

// scanPosts is a helper method that scans database rows into post entities
// This method handles the repetitive task of scanning post data from database rows.
//
// Parameters:
//   - rows: database rows containing post data
//
// Returns:
//   - []*entities.Post: slice of scanned post entities
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresPostRepository) scanPosts(rows pgx.Rows) ([]*entities.Post, error) {
	var posts []*entities.Post

	for rows.Next() {
		var post entities.Post
		var contentRaw *string
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Subtitle,
			&post.Description,
			&post.Type,
			&post.Language,
			&post.Layout,
			&post.IsSticky,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.DeletedAt,
			&contentRaw,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		// Set content from the joined content table
		if contentRaw != nil {
			post.Content = *contentRaw
		}

		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over post rows: %w", err)
	}

	return posts, nil
}
