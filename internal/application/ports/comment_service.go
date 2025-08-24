package ports

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// CommentService defines the application service interface for comment operations
type CommentService interface {
	CreateComment(ctx context.Context, comment *entities.Comment) (*entities.Comment, error)
	GetCommentByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error)
	GetCommentsByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error)
	GetCommentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error)
	GetCommentReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error)
	GetAllComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error)
	GetApprovedComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error)
	GetPendingComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error)
	UpdateComment(ctx context.Context, comment *entities.Comment) (*entities.Comment, error)
	DeleteComment(ctx context.Context, id uuid.UUID) error
	ApproveComment(ctx context.Context, id uuid.UUID) error
	RejectComment(ctx context.Context, id uuid.UUID) error
	GetCommentCount(ctx context.Context) (int64, error)
	GetCommentCountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
	GetCommentCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	GetPendingCommentCount(ctx context.Context) (int64, error)
}
