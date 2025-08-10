package adapters

import (
	"context"
	"strings"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresContentRepository struct {
	repo repository.ContentRepository
}

func NewPostgresContentRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.ContentRepository {

	return &PostgresContentRepository{
		repo: repository.NewContentRepository(db, redisClient),
	}
}

func (r *PostgresContentRepository) Create(ctx context.Context, content *entities.Content) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Create(ctx, content)
}

func (r *PostgresContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByID(ctx, id)
}

func (r *PostgresContentRepository) GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByModelTypeAndID(ctx, modelType, modelID)
}

func (r *PostgresContentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetAll(ctx, limit, offset)
}

func (r *PostgresContentRepository) Update(ctx context.Context, content *entities.Content) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Update(ctx, content)
}

func (r *PostgresContentRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	// This method signature doesn't match the repository interface
	// We need to implement soft delete logic here
	content, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Mark as deleted
	content.MarkAsDeleted(deletedBy)
	return r.repo.Update(ctx, content)
}

func (r *PostgresContentRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Delete(ctx, id)
}

func (r *PostgresContentRepository) Restore(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	// This method is not available in the repository interface
	// We need to implement it by getting the content and updating it
	content, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Restore the content
	content.Restore(updatedBy)
	return r.repo.Update(ctx, content)
}

func (r *PostgresContentRepository) Count(ctx context.Context) (int64, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Count(ctx)
}

func (r *PostgresContentRepository) CountByModelType(ctx context.Context, modelType string) (int64, error) {
	// This method is not available in the repository interface
	// We need to implement it by counting filtered results
	allContents, err := r.repo.GetAll(ctx, 1000, 0) // Get a large number to count
	if err != nil {
		return 0, err
	}

	var count int64
	for _, content := range allContents {
		if content.ModelType == modelType {
			count++
		}
	}

	return count, nil
}

func (r *PostgresContentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	// This method is not available in the repository interface
	// We need to implement it by searching through all contents
	allContents, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var searchResults []*entities.Content
	for _, content := range allContents {
		// Simple text search in content_raw and content_html
		if strings.Contains(strings.ToLower(content.ContentRaw), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(content.ContentHTML), strings.ToLower(query)) {
			searchResults = append(searchResults, content)
		}
	}

	return searchResults, nil
}
