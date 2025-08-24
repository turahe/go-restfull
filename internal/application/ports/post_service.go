package ports

import (
	"context"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// PostService defines the application service interface for post operations
type PostService interface {
	// CreatePost creates a new post
	CreatePost(ctx context.Context, title, slug, subtitle, description, language, layout, content string, isSticky bool, publishedAt *time.Time) (*entities.Post, error)

	// GetPostByID retrieves a post by ID
	GetPostByID(ctx context.Context, id uuid.UUID) (*entities.Post, error)

	// GetPostBySlug retrieves a post by slug
	GetPostBySlug(ctx context.Context, slug string) (*entities.Post, error)

	// GetPostsByAuthor retrieves posts by author ID
	GetPostsByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.Post, error)

	// GetAllPosts retrieves all posts with pagination
	GetAllPosts(ctx context.Context, limit, offset int) ([]*entities.Post, error)

	// GetPublishedPosts retrieves published posts with pagination
	GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entities.Post, error)

	// SearchPosts searches posts by query
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error)

	// GetPostsWithPagination retrieves posts with pagination and returns total count
	GetPostsWithPagination(ctx context.Context, page, perPage int, search, status string) ([]*entities.Post, int64, error)

	// GetPostsCount returns total count of posts (for pagination)
	GetPostsCount(ctx context.Context, search, status string) (int64, error)

	// UpdatePost updates post information
	UpdatePost(ctx context.Context, id uuid.UUID, title, slug, subtitle, description, language, layout string, isSticky bool, publishedAt *time.Time) (*entities.Post, error)

	// DeletePost soft deletes a post
	DeletePost(ctx context.Context, id uuid.UUID) error

	// PublishPost publishes a post
	PublishPost(ctx context.Context, id uuid.UUID) error

	// UnpublishPost unpublishes a post
	UnpublishPost(ctx context.Context, id uuid.UUID) error
}
