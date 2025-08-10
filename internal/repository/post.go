package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostRepository interface {
	Create(ctx context.Context, post *entities.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Post, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Post, error)
	GetPublished(ctx context.Context, limit, offset int) ([]*entities.Post, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error)
	Update(ctx context.Context, post *entities.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	Publish(ctx context.Context, id uuid.UUID) error
	Unpublish(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountPublished(ctx context.Context) (int64, error)
}

type PostRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) PostRepository {
	return &PostRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *PostRepositoryImpl) Create(ctx context.Context, post *entities.Post) error {
	query := `INSERT INTO posts (id, title, slug, subtitle, description, language, layout, is_sticky, published_at, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pgxPool.Exec(ctx, query,
		post.ID.String(), post.Title, post.Slug, post.Subtitle, post.Description, post.Language, post.Layout, post.IsSticky, post.PublishedAt, post.CreatedBy, post.UpdatedBy, post.CreatedAt, post.UpdatedAt)
	return err
}

func (r *PostRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	query := `SELECT id, title, slug, content, status, author_id, published_at, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM posts WHERE id = $1 AND deleted_at IS NULL`

	var post entities.Post
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&post.ID, &post.Title, &post.Slug, &post.Subtitle, &post.Description, &post.Language, &post.Layout, &post.IsSticky, &post.PublishedAt, &post.CreatedBy, &post.UpdatedBy, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entities.Post, error) {
	query := `SELECT id, title, slug, content, status, author_id, published_at, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM posts WHERE slug = $1 AND deleted_at IS NULL`

	var post entities.Post
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(
		&post.ID, &post.Title, &post.Slug, &post.Subtitle, &post.Description, &post.Language, &post.Layout, &post.IsSticky, &post.PublishedAt, &post.CreatedBy, &post.UpdatedBy, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	query := `SELECT id, title, slug, subtitle, description, language, layout, is_sticky, published_at, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM posts WHERE deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*entities.Post
	for rows.Next() {
		post, err := r.scanPostRow(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepositoryImpl) GetPublished(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	query := `SELECT id, title, slug, subtitle, description, language, layout, is_sticky, published_at, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM posts WHERE published_at IS NOT NULL AND deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*entities.Post
	for rows.Next() {
		post, err := r.scanPostRow(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	searchQuery := `SELECT id, title, slug, subtitle, description, language, layout, is_sticky, published_at, created_by, updated_by, created_at, updated_at, deleted_at
					FROM posts WHERE deleted_at IS NULL AND
					(title ILIKE $1 OR slug ILIKE $1 OR subtitle ILIKE $1 OR description ILIKE $1)
					ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	searchTerm := "%" + query + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*entities.Post
	for rows.Next() {
		post, err := r.scanPostRow(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepositoryImpl) Update(ctx context.Context, post *entities.Post) error {
	query := `UPDATE posts SET title = $1, slug = $2, subtitle = $3, description = $4, language = $5, layout = $6, is_sticky = $7, published_at = $8, updated_by = $9, updated_at = $10
			  WHERE id = $11 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, post.Title, post.Slug, post.Subtitle, post.Description, post.Language, post.Layout, post.IsSticky, post.PublishedAt, post.UpdatedBy,
		post.UpdatedAt, post.ID.String())
	return err
}

func (r *PostRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE posts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *PostRepositoryImpl) Publish(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE posts SET published_at = NOW(), updated_at = NOW(), updated_by = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *PostRepositoryImpl) Unpublish(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE posts SET published_at = NULL, updated_at = NOW(), updated_by = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *PostRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

func (r *PostRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *PostRepositoryImpl) CountPublished(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM posts WHERE published_at IS NOT NULL AND deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// scanPostRow is a helper function to scan a post row from database
func (r *PostRepositoryImpl) scanPostRow(rows pgx.Rows) (*entities.Post, error) {
	var post entities.Post
	err := rows.Scan(
		&post.ID, &post.Title, &post.Slug, &post.Subtitle, &post.Description, &post.Language, &post.Layout, &post.IsSticky, &post.PublishedAt, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}
