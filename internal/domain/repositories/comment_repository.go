package repositories

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// CommentRepository defines the interface for comment data access
type CommentRepository interface {
	Create(ctx context.Context, comment *entities.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error)
	GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error)
	GetReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Comment, error)
	GetApproved(ctx context.Context, limit, offset int) ([]*entities.Comment, error)
	GetPending(ctx context.Context, limit, offset int) ([]*entities.Comment, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Comment, error)
	Update(ctx context.Context, comment *entities.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	CountPending(ctx context.Context) (int64, error)
}
