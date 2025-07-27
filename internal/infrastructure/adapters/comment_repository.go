package adapters

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"
	"webapi/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresCommentRepository struct {
	repo repository.CommentRepositoryInterface
}

func NewPostgresCommentRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.CommentRepository {
	return &PostgresCommentRepository{
		repo: repository.NewCommentRepository(db, redisClient),
	}
}

func (r *PostgresCommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	// Convert domain entity to model
	commentModel := &model.Comment{
		ID:             comment.ID,
		ModelType:      "post", // Default model type
		ModelID:        comment.PostID,
		Title:          "", // Not available in domain entity
		Status:         comment.Status,
		ParentID:       comment.ParentID,
		RecordLeft:     0, // Will be set by repository
		RecordRight:    0, // Will be set by repository
		RecordDepth:    0, // Will be set by repository
		RecordOrdering: 0, // Will be set by repository
		CreatedBy:      comment.UserID,
		UpdatedBy:      comment.UserID,
		DeletedBy:      uuid.Nil,
		DeletedAt:      comment.DeletedAt,
		CreatedAt:      comment.CreatedAt,
		UpdatedAt:      comment.UpdatedAt,
	}

	return r.repo.CreateComment(ctx, commentModel)
}

func (r *PostgresCommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	commentModel, err := r.repo.GetCommentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert model to domain entity
	comment := &entities.Comment{
		ID:        commentModel.ID,
		Content:   "", // Not available in model, would need to get from contents
		PostID:    commentModel.ModelID,
		UserID:    commentModel.CreatedBy,
		ParentID:  commentModel.ParentID,
		Status:    commentModel.Status,
		CreatedAt: commentModel.CreatedAt,
		UpdatedAt: commentModel.UpdatedAt,
		DeletedAt: commentModel.DeletedAt,
	}

	return comment, nil
}

func (r *PostgresCommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning empty slice as the existing repository doesn't have this method
	return []*entities.Comment{}, nil
}

func (r *PostgresCommentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning empty slice as the existing repository doesn't have this method
	return []*entities.Comment{}, nil
}

func (r *PostgresCommentRepository) GetReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning empty slice as the existing repository doesn't have this method
	return []*entities.Comment{}, nil
}

func (r *PostgresCommentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	commentModels, err := r.repo.GetAllComments(ctx)
	if err != nil {
		return nil, err
	}

	// Convert models to domain entities
	var comments []*entities.Comment
	for _, commentModel := range commentModels {
		comments = append(comments, &entities.Comment{
			ID:        commentModel.ID,
			Content:   "", // Not available in model, would need to get from contents
			PostID:    commentModel.ModelID,
			UserID:    commentModel.CreatedBy,
			ParentID:  commentModel.ParentID,
			Status:    commentModel.Status,
			CreatedAt: commentModel.CreatedAt,
			UpdatedAt: commentModel.UpdatedAt,
			DeletedAt: commentModel.DeletedAt,
		})
	}

	return comments, nil
}

func (r *PostgresCommentRepository) GetApproved(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning empty slice as the existing repository doesn't have this method
	return []*entities.Comment{}, nil
}

func (r *PostgresCommentRepository) GetPending(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning empty slice as the existing repository doesn't have this method
	return []*entities.Comment{}, nil
}

func (r *PostgresCommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	// Convert domain entity to model
	commentModel := &model.Comment{
		ID:             comment.ID,
		ModelType:      "post", // Default model type
		ModelID:        comment.PostID,
		Title:          "", // Not available in domain entity
		Status:         comment.Status,
		ParentID:       comment.ParentID,
		RecordLeft:     0, // Will be set by repository
		RecordRight:    0, // Will be set by repository
		RecordDepth:    0, // Will be set by repository
		RecordOrdering: 0, // Will be set by repository
		CreatedBy:      comment.UserID,
		UpdatedBy:      comment.UserID,
		DeletedBy:      uuid.Nil,
		DeletedAt:      comment.DeletedAt,
		CreatedAt:      comment.CreatedAt,
		UpdatedAt:      comment.UpdatedAt,
	}

	return r.repo.UpdateComment(ctx, commentModel)
}

func (r *PostgresCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.DeleteComment(ctx, id)
}

func (r *PostgresCommentRepository) Approve(ctx context.Context, id uuid.UUID) error {
	// This would need to be implemented based on your specific requirements
	// For now, returning nil as the existing repository doesn't have this method
	return nil
}

func (r *PostgresCommentRepository) Reject(ctx context.Context, id uuid.UUID) error {
	// This would need to be implemented based on your specific requirements
	// For now, returning nil as the existing repository doesn't have this method
	return nil
}

func (r *PostgresCommentRepository) Count(ctx context.Context) (int64, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning 0 as the existing repository doesn't have this method
	return 0, nil
}

func (r *PostgresCommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning 0 as the existing repository doesn't have this method
	return 0, nil
}

func (r *PostgresCommentRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning 0 as the existing repository doesn't have this method
	return 0, nil
}

func (r *PostgresCommentRepository) CountPending(ctx context.Context) (int64, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning 0 as the existing repository doesn't have this method
	return 0, nil
}
