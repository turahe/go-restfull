package ports

import (
	"context"
	"mime/multipart"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// MediaService defines the application service interface for media operations
type MediaService interface {
	UploadMedia(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*entities.Media, error)
	GetMediaByID(ctx context.Context, id uuid.UUID) (*entities.Media, error)
	GetMediaByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error)
	GetAvatarByUserID(ctx context.Context, userID uuid.UUID) (*entities.Media, error)
	GetMediaByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string) (*entities.Media, error)
	GetAllMediaByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string, limit, offset int) ([]*entities.Media, error)
	GetAllMedia(ctx context.Context, limit, offset int) ([]*entities.Media, error)
	SearchMedia(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error)
	UpdateMedia(ctx context.Context, id uuid.UUID, fileName, originalName, mimeType, path, url string, size int64) (*entities.Media, error)
	DeleteMedia(ctx context.Context, id uuid.UUID) error
	GetMediaCount(ctx context.Context) (int64, error)
	GetMediaCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	AttachMediaToEntity(ctx context.Context, mediaID uuid.UUID, mediableID uuid.UUID, mediableType, group string) error
}
