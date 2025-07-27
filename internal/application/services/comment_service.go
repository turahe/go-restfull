package services

import (
	"context"

	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"

	"github.com/google/uuid"
)

// commentService implements CommentService interface
type commentService struct {
	commentRepository repositories.CommentRepository
}

// NewCommentService creates a new comment service
func NewCommentService(commentRepository repositories.CommentRepository) ports.CommentService {
	return &commentService{
		commentRepository: commentRepository,
	}
}

// CreateComment creates a new comment
func (s *commentService) CreateComment(ctx context.Context, content string, postID, userID uuid.UUID, parentID *uuid.UUID, status string) (*entities.Comment, error) {
	comment, err := entities.NewComment(content, postID, userID, parentID, status)
	if err != nil {
		return nil, err
	}

	err = s.commentRepository.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetCommentByID retrieves comment by ID
func (s *commentService) GetCommentByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	return s.commentRepository.GetByID(ctx, id)
}

// GetCommentsByPostID retrieves comments by post ID
func (s *commentService) GetCommentsByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetByPostID(ctx, postID, limit, offset)
}

// GetCommentsByUserID retrieves comments by user ID
func (s *commentService) GetCommentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetByUserID(ctx, userID, limit, offset)
}

// GetCommentReplies retrieves replies to a comment
func (s *commentService) GetCommentReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetReplies(ctx, parentID, limit, offset)
}

// GetAllComments retrieves all comments with pagination
func (s *commentService) GetAllComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetAll(ctx, limit, offset)
}

// GetApprovedComments retrieves approved comments
func (s *commentService) GetApprovedComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetApproved(ctx, limit, offset)
}

// GetPendingComments retrieves pending comments
func (s *commentService) GetPendingComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetPending(ctx, limit, offset)
}

// UpdateComment updates comment information
func (s *commentService) UpdateComment(ctx context.Context, id uuid.UUID, content, status string) (*entities.Comment, error) {
	comment, err := s.commentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = comment.UpdateComment(content, status)
	if err != nil {
		return nil, err
	}

	err = s.commentRepository.Update(ctx, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment deletes comment
func (s *commentService) DeleteComment(ctx context.Context, id uuid.UUID) error {
	return s.commentRepository.Delete(ctx, id)
}

// ApproveComment approves a comment
func (s *commentService) ApproveComment(ctx context.Context, id uuid.UUID) error {
	return s.commentRepository.Approve(ctx, id)
}

// RejectComment rejects a comment
func (s *commentService) RejectComment(ctx context.Context, id uuid.UUID) error {
	return s.commentRepository.Reject(ctx, id)
}

// GetCommentCount returns the total number of comments
func (s *commentService) GetCommentCount(ctx context.Context) (int64, error) {
	return s.commentRepository.Count(ctx)
}

// GetCommentCountByPostID returns the total number of comments by post ID
func (s *commentService) GetCommentCountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	return s.commentRepository.CountByPostID(ctx, postID)
}

// GetCommentCountByUserID returns the total number of comments by user ID
func (s *commentService) GetCommentCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.commentRepository.CountByUserID(ctx, userID)
}

// GetPendingCommentCount returns the total number of pending comments
func (s *commentService) GetPendingCommentCount(ctx context.Context) (int64, error) {
	return s.commentRepository.CountPending(ctx)
}
