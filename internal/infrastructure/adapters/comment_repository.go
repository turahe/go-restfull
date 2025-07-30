package adapters

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresCommentRepository struct {
	repo repository.CommentRepository
}

func NewPostgresCommentRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.CommentRepository {
	return &PostgresCommentRepository{
		repo: repository.NewCommentRepository(db, redisClient),
	}
}

func (r *PostgresCommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Create(ctx, comment)
}

func (r *PostgresCommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByID(ctx, id)
}

func (r *PostgresCommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetByPostID(ctx, postID, limit, offset)
}

func (r *PostgresCommentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allComments, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var userComments []*entities.Comment
	for _, comment := range allComments {
		if comment.UserID == userID {
			userComments = append(userComments, comment)
		}
	}

	return userComments, nil
}

func (r *PostgresCommentRepository) GetReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allComments, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var replies []*entities.Comment
	for _, comment := range allComments {
		if comment.ParentID != nil && *comment.ParentID == parentID {
			replies = append(replies, comment)
		}
	}

	return replies, nil
}

func (r *PostgresCommentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.GetAll(ctx, limit, offset)
}

func (r *PostgresCommentRepository) GetApproved(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allComments, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var approvedComments []*entities.Comment
	for _, comment := range allComments {
		if comment.Status == "approved" {
			approvedComments = append(approvedComments, comment)
		}
	}

	return approvedComments, nil
}

func (r *PostgresCommentRepository) GetPending(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allComments, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var pendingComments []*entities.Comment
	for _, comment := range allComments {
		if comment.Status == "pending" {
			pendingComments = append(pendingComments, comment)
		}
	}

	return pendingComments, nil
}

func (r *PostgresCommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Update(ctx, comment)
}

func (r *PostgresCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Delete(ctx, id)
}

func (r *PostgresCommentRepository) Approve(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Approve(ctx, id)
}

func (r *PostgresCommentRepository) Reject(ctx context.Context, id uuid.UUID) error {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Reject(ctx, id)
}

func (r *PostgresCommentRepository) Count(ctx context.Context) (int64, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.Count(ctx)
}

func (r *PostgresCommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	// The repository already works with entities, so we can pass through directly
	return r.repo.CountByPostID(ctx, postID)
}

func (r *PostgresCommentRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	// This method is not available in the repository interface
	// We need to implement it by counting filtered results
	allComments, err := r.repo.GetAll(ctx, 1000, 0) // Get a large number to count
	if err != nil {
		return 0, err
	}

	var count int64
	for _, comment := range allComments {
		if comment.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *PostgresCommentRepository) CountPending(ctx context.Context) (int64, error) {
	// This method is not available in the repository interface
	// We need to implement it by counting filtered results
	allComments, err := r.repo.GetAll(ctx, 1000, 0) // Get a large number to count
	if err != nil {
		return 0, err
	}

	var count int64
	for _, comment := range allComments {
		if comment.Status == "pending" {
			count++
		}
	}

	return count, nil
}
