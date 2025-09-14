package repositories

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	Create(ctx context.Context, media *entities.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error)
	GetAvatarByUserID(ctx context.Context, userID uuid.UUID) (*entities.Media, error)
	GetByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string) (*entities.Media, error)
	GetAllByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string, limit, offset int) ([]*entities.Media, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error)
	Update(ctx context.Context, media *entities.Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	AttachMediaToEntity(ctx context.Context, mediaID uuid.UUID, mediableID uuid.UUID, mediableType, group string) error
}
