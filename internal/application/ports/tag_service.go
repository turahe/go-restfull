package ports

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// TagService defines the application service interface for tag operations
type TagService interface {
	CreateTag(ctx context.Context, name, slug, description, color string) (*entities.Tag, error)
	GetTagByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error)
	GetTagBySlug(ctx context.Context, slug string) (*entities.Tag, error)
	GetAllTags(ctx context.Context, limit, offset int) ([]*entities.Tag, error)
	SearchTags(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error)
	UpdateTag(ctx context.Context, id uuid.UUID, name, slug, description, color string) (*entities.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
	GetTagCount(ctx context.Context) (int64, error)
}
