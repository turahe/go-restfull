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

type PostgresTagRepository struct {
	repo repository.TagRepository
}

func NewPostgresTagRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.TagRepository {
	return &PostgresTagRepository{
		repo: repository.NewTagRepository(db, redisClient),
	}
}

func (r *PostgresTagRepository) Create(ctx context.Context, tag *entities.Tag) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Create(ctx, tag)
}

func (r *PostgresTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByID(ctx, id)
}

func (r *PostgresTagRepository) GetBySlug(ctx context.Context, slug string) (*entities.Tag, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allTags, err := r.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, tag := range allTags {
		if tag.Slug == slug {
			return tag, nil
		}
	}

	return nil, nil // Not found
}

func (r *PostgresTagRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Tag, error) {
	// The repository method doesn't take limit and offset parameters
	// We need to get all tags and then apply pagination
	allTags, err := r.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Apply pagination manually
	start := offset
	end := offset + limit
	if start >= len(allTags) {
		return []*entities.Tag{}, nil
	}
	if end > len(allTags) {
		end = len(allTags)
	}

	return allTags[start:end], nil
}

func (r *PostgresTagRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	// This method is not available in the repository interface
	// We need to implement it by searching through all tags
	allTags, err := r.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var searchResults []*entities.Tag
	queryLower := strings.ToLower(query)
	for _, tag := range allTags {
		if strings.Contains(strings.ToLower(tag.Name), queryLower) ||
			strings.Contains(strings.ToLower(tag.Slug), queryLower) ||
			strings.Contains(strings.ToLower(tag.Description), queryLower) {
			searchResults = append(searchResults, tag)
		}
	}

	// Apply pagination to search results
	start := offset
	end := offset + limit
	if start >= len(searchResults) {
		return []*entities.Tag{}, nil
	}
	if end > len(searchResults) {
		end = len(searchResults)
	}

	return searchResults[start:end], nil
}

func (r *PostgresTagRepository) Update(ctx context.Context, tag *entities.Tag) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Update(ctx, tag)
}

func (r *PostgresTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Delete(ctx, id)
}

func (r *PostgresTagRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	// This method is not available in the repository interface
	// We need to implement it by checking if a tag with the slug exists
	allTags, err := r.repo.GetAll(ctx)
	if err != nil {
		return false, err
	}

	for _, tag := range allTags {
		if tag.Slug == slug {
			return true, nil
		}
	}

	return false, nil
}

func (r *PostgresTagRepository) Count(ctx context.Context) (int64, error) {
	// This method is not available in the repository interface
	// We need to implement it by counting all tags
	allTags, err := r.repo.GetAll(ctx)
	if err != nil {
		return 0, err
	}

	return int64(len(allTags)), nil
}
