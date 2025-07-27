package repositories

import (
	"context"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// PostRepository defines the interface for post data access
type PostRepository interface {
	// Create creates a new post
	Create(ctx context.Context, post *entities.Post) error

	// GetByID retrieves a post by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error)

	// GetBySlug retrieves a post by slug
	GetBySlug(ctx context.Context, slug string) (*entities.Post, error)

	// GetByAuthor retrieves posts by author ID
	GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.Post, error)

	// GetAll retrieves all posts with optional pagination
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Post, error)

	// GetPublished retrieves published posts with pagination
	GetPublished(ctx context.Context, limit, offset int) ([]*entities.Post, error)

	// Search searches posts by query
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error)

	// Update updates an existing post
	Update(ctx context.Context, post *entities.Post) error

	// Delete soft deletes a post
	Delete(ctx context.Context, id uuid.UUID) error

	// Publish publishes a post
	Publish(ctx context.Context, id uuid.UUID) error

	// Unpublish unpublishes a post
	Unpublish(ctx context.Context, id uuid.UUID) error

	// Count returns the total number of posts
	Count(ctx context.Context) (int64, error)

	// CountPublished returns the total number of published posts
	CountPublished(ctx context.Context) (int64, error)

	// CountBySearch returns the total number of posts matching the search query
	CountBySearch(ctx context.Context, query string) (int64, error)

	// CountBySearchPublished returns the total number of published posts matching the search query
	CountBySearchPublished(ctx context.Context, query string) (int64, error)

	// SearchPublished searches published posts by query
	SearchPublished(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error)
}
