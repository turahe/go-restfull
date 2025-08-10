package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresPostRepository is an adapter that implements the domain PostRepository interface
// by delegating to the concrete repository implementation
type PostgresPostRepository struct {
	*BaseTransactionalRepository
	repo repositories.PostRepository
}

// NewPostgresPostRepository creates a new PostgreSQL post repository adapter
func NewPostgresPostRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.PostRepository {
	return &PostgresPostRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		repo:                        repository.NewPostgresPostRepository(db, redisClient),
	}
}

// Create delegates to the underlying repository implementation
func (r *PostgresPostRepository) Create(ctx context.Context, post *entities.Post) error {
	return r.repo.Create(ctx, post)
}

// GetByID delegates to the underlying repository implementation
func (r *PostgresPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	return r.repo.GetByID(ctx, id)
}

// GetBySlug delegates to the underlying repository implementation
func (r *PostgresPostRepository) GetBySlug(ctx context.Context, slug string) (*entities.Post, error) {
	return r.repo.GetBySlug(ctx, slug)
}

// GetByAuthor delegates to the underlying repository implementation
func (r *PostgresPostRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.Post, error) {
	return r.repo.GetByAuthor(ctx, authorID, limit, offset)
}

// GetAll delegates to the underlying repository implementation
func (r *PostgresPostRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	return r.repo.GetAll(ctx, limit, offset)
}

// GetPublished delegates to the underlying repository implementation
func (r *PostgresPostRepository) GetPublished(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	return r.repo.GetPublished(ctx, limit, offset)
}

// Search delegates to the underlying repository implementation
func (r *PostgresPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	return r.repo.Search(ctx, query, limit, offset)
}

// Update delegates to the underlying repository implementation
func (r *PostgresPostRepository) Update(ctx context.Context, post *entities.Post) error {
	return r.repo.Update(ctx, post)
}

// Delete delegates to the underlying repository implementation
func (r *PostgresPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

// Publish delegates to the underlying repository implementation
func (r *PostgresPostRepository) Publish(ctx context.Context, id uuid.UUID) error {
	return r.repo.Publish(ctx, id)
}

// Unpublish delegates to the underlying repository implementation
func (r *PostgresPostRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	return r.repo.Unpublish(ctx, id)
}

// Count delegates to the underlying repository implementation
func (r *PostgresPostRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

// CountPublished delegates to the underlying repository implementation
func (r *PostgresPostRepository) CountPublished(ctx context.Context) (int64, error) {
	return r.repo.CountPublished(ctx)
}

// CountBySearch delegates to the underlying repository implementation
func (r *PostgresPostRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
	return r.repo.CountBySearch(ctx, query)
}

// CountBySearchPublished delegates to the underlying repository implementation
func (r *PostgresPostRepository) CountBySearchPublished(ctx context.Context, query string) (int64, error) {
	return r.repo.CountBySearchPublished(ctx, query)
}

// SearchPublished delegates to the underlying repository implementation
func (r *PostgresPostRepository) SearchPublished(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	return r.repo.SearchPublished(ctx, query, limit, offset)
}
