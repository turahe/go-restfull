package repositories

import (
	"context"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// TagRepository defines the interface for tag data access
type TagRepository interface {
	Create(ctx context.Context, tag *entities.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Tag, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Tag, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error)
	Update(ctx context.Context, tag *entities.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
}
