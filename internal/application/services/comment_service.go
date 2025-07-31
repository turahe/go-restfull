// Package services provides application-level business logic for comment management.
// This package contains the comment service implementation that handles comment creation,
// moderation, retrieval, and hierarchical comment structures while enforcing business rules.
package services

import (
	"context"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// commentService implements the CommentService interface and provides comprehensive
// comment management functionality. It handles comment creation, moderation, retrieval,
// hierarchical structures (replies), and status management while enforcing business rules.
type commentService struct {
	commentRepository repositories.CommentRepository
}

// NewCommentService creates a new comment service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - commentRepository: Repository interface for comment data access operations
//
// Returns:
//   - ports.CommentService: The comment service interface implementation
func NewCommentService(commentRepository repositories.CommentRepository) ports.CommentService {
	return &commentService{
		commentRepository: commentRepository,
	}
}

// CreateComment creates a new comment for a specific post with optional parent comment.
// This method enforces business rules for comment creation and supports hierarchical
// comment structures (replies to comments).
//
// Business Rules:
//   - Comment content must be provided and validated
//   - Post ID must reference an existing post
//   - User ID must reference an existing user
//   - Parent ID is optional for top-level comments
//   - Status determines comment visibility and moderation state
//
// Parameters:
//   - ctx: Context for the operation
//   - content: The comment text content
//   - postID: UUID of the post this comment belongs to
//   - userID: UUID of the user creating the comment
//   - parentID: Optional UUID of the parent comment (for replies)
//   - status: Comment status (pending, approved, rejected)
//
// Returns:
//   - *entities.Comment: The created comment entity
//   - error: Any error that occurred during the operation
func (s *commentService) CreateComment(ctx context.Context, content string, postID, userID uuid.UUID, parentID *uuid.UUID, status string) (*entities.Comment, error) {
	// Create new comment entity with the provided data
	comment, err := entities.NewComment(content, postID, userID, parentID, status)
	if err != nil {
		return nil, err
	}

	// Persist the comment to the repository
	err = s.commentRepository.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetCommentByID retrieves a comment by its unique identifier.
// This method includes soft delete checking to ensure deleted comments
// are not returned to the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the comment to retrieve
//
// Returns:
//   - *entities.Comment: The comment entity if found
//   - error: Error if comment not found or other issues occur
func (s *commentService) GetCommentByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	return s.commentRepository.GetByID(ctx, id)
}

// GetCommentsByPostID retrieves all comments for a specific post with pagination.
// This method supports comment moderation by returning comments based on their status.
//
// Parameters:
//   - ctx: Context for the operation
//   - postID: UUID of the post to get comments for
//   - limit: Maximum number of comments to return
//   - offset: Number of comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of comments for the post
//   - error: Any error that occurred during the operation
func (s *commentService) GetCommentsByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetByPostID(ctx, postID, limit, offset)
}

// GetCommentsByUserID retrieves all comments created by a specific user with pagination.
// This method is useful for user profile pages and comment history.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to get comments for
//   - limit: Maximum number of comments to return
//   - offset: Number of comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of comments by the user
//   - error: Any error that occurred during the operation
func (s *commentService) GetCommentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetByUserID(ctx, userID, limit, offset)
}

// GetCommentReplies retrieves all replies to a specific comment with pagination.
// This method supports hierarchical comment structures by returning child comments.
//
// Parameters:
//   - ctx: Context for the operation
//   - parentID: UUID of the parent comment
//   - limit: Maximum number of replies to return
//   - offset: Number of replies to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of reply comments
//   - error: Any error that occurred during the operation
func (s *commentService) GetCommentReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetReplies(ctx, parentID, limit, offset)
}

// GetAllComments retrieves all comments in the system with pagination.
// This method is useful for administrative purposes and system-wide comment management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of comments to return
//   - offset: Number of comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of all comments
//   - error: Any error that occurred during the operation
func (s *commentService) GetAllComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetAll(ctx, limit, offset)
}

// GetApprovedComments retrieves only approved comments with pagination.
// This method is useful for displaying public comments that have passed moderation.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of approved comments to return
//   - offset: Number of approved comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of approved comments
//   - error: Any error that occurred during the operation
func (s *commentService) GetApprovedComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetApproved(ctx, limit, offset)
}

// GetPendingComments retrieves only pending comments with pagination.
// This method is useful for moderation workflows and administrative review.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of pending comments to return
//   - offset: Number of pending comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of pending comments
//   - error: Any error that occurred during the operation
func (s *commentService) GetPendingComments(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	return s.commentRepository.GetPending(ctx, limit, offset)
}

// UpdateComment updates an existing comment's content and status.
// This method enforces business rules and maintains data integrity
// during the update process.
//
// Business Rules:
//   - Comment must exist and not be deleted
//   - Content must be provided and validated
//   - Status changes are tracked for moderation purposes
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the comment to update
//   - content: Updated comment text content
//   - status: Updated comment status
//
// Returns:
//   - *entities.Comment: The updated comment entity
//   - error: Any error that occurred during the operation
func (s *commentService) UpdateComment(ctx context.Context, id uuid.UUID, content, status string) (*entities.Comment, error) {
	// Retrieve existing comment to ensure it exists and is not deleted
	comment, err := s.commentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the comment entity with new content and status
	err = comment.UpdateComment(content, status)
	if err != nil {
		return nil, err
	}

	// Persist the updated comment to the repository
	err = s.commentRepository.Update(ctx, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment performs a soft delete of a comment by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the comment to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *commentService) DeleteComment(ctx context.Context, id uuid.UUID) error {
	return s.commentRepository.Delete(ctx, id)
}

// ApproveComment approves a pending comment, making it visible to the public.
// This method is part of the comment moderation workflow.
//
// Business Rules:
//   - Comment must exist and be in pending status
//   - Approval changes the comment status to approved
//   - Approved comments become visible to all users
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the comment to approve
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *commentService) ApproveComment(ctx context.Context, id uuid.UUID) error {
	return s.commentRepository.Approve(ctx, id)
}

// RejectComment rejects a pending comment, preventing it from being displayed.
// This method is part of the comment moderation workflow.
//
// Business Rules:
//   - Comment must exist and be in pending status
//   - Rejection changes the comment status to rejected
//   - Rejected comments are not visible to the public
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the comment to reject
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *commentService) RejectComment(ctx context.Context, id uuid.UUID) error {
	return s.commentRepository.Reject(ctx, id)
}

// GetCommentCount returns the total number of comments in the system.
// This method is useful for statistics and administrative reporting.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of comments
//   - error: Any error that occurred during the operation
func (s *commentService) GetCommentCount(ctx context.Context) (int64, error) {
	return s.commentRepository.Count(ctx)
}

// GetCommentCountByPostID returns the total number of comments for a specific post.
// This method is useful for displaying comment counts on post listings.
//
// Parameters:
//   - ctx: Context for the operation
//   - postID: UUID of the post to count comments for
//
// Returns:
//   - int64: Total count of comments for the post
//   - error: Any error that occurred during the operation
func (s *commentService) GetCommentCountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	return s.commentRepository.CountByPostID(ctx, postID)
}

// GetCommentCountByUserID returns the total number of comments created by a specific user.
// This method is useful for user profile statistics and activity tracking.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to count comments for
//
// Returns:
//   - int64: Total count of comments by the user
//   - error: Any error that occurred during the operation
func (s *commentService) GetCommentCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.commentRepository.CountByUserID(ctx, userID)
}

// GetPendingCommentCount returns the total number of pending comments in the system.
// This method is useful for moderation dashboards and administrative alerts.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of pending comments
//   - error: Any error that occurred during the operation
func (s *commentService) GetPendingCommentCount(ctx context.Context) (int64, error) {
	return s.commentRepository.CountPending(ctx)
}
