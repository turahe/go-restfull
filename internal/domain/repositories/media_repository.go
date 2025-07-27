package repositories

import (
	"context"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	Create(ctx context.Context, media *entities.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error)
	Update(ctx context.Context, media *entities.Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
