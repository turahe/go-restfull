// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreateCommentRequest represents the request for creating a new comment entity.
// This struct defines the required and optional fields for comment creation,
// including content, post association, and optional parent comment for replies.
// Comments support hierarchical structures through parent_id for threaded discussions.
type CreateCommentRequest struct {
	// Content is the text content of the comment (required, 1-1000 characters)
	Content string `json:"content" validate:"required,min=1,max=1000"`
	// PostID is the UUID of the post this comment belongs to (required)
	PostID uuid.UUID `json:"post_id" validate:"required"`
	// ParentID is the UUID of the parent comment for replies (optional, nil for top-level comments)
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// ToEntity transforms the CreateCommentRequest to a Comment domain entity.
// This method creates a new comment with default values and associates it
// with the specified post and optional parent comment.
//
// Parameters:
//   - authorID: The UUID of the user creating the comment
//
// Returns:
//   - *entities.Comment: The created comment entity with default status and metadata
func (r *CreateCommentRequest) ToEntity(authorID uuid.UUID) *entities.Comment {
	return &entities.Comment{
		ID:        uuid.New(),
		ModelID:   r.PostID,
		ModelType: "post",
		ParentID:  r.ParentID,
		CreatedBy: authorID,
		Status:    entities.CommentStatusPending, // Default to pending for moderation
	}
}

// UpdateCommentRequest represents the request for updating an existing comment entity.
// This struct allows updating the comment content while preserving other metadata.
// Currently limited to content updates based on the Comment entity structure.
type UpdateCommentRequest struct {
	// Content is the updated text content of the comment (required, 1-1000 characters)
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// ToEntity transforms the UpdateCommentRequest to update an existing Comment entity.
// This method is designed for partial updates, but currently limited by the Comment
// entity structure which may not include a Content field.
//
// Note: The current Comment entity structure may need to be updated to support
// content updates, or this method may need to be modified based on the actual
// entity design.
//
// Parameters:
//   - existingComment: The existing comment entity to update
//
// Returns:
//   - *entities.Comment: The comment entity (currently unchanged due to structure limitations)
func (r *UpdateCommentRequest) ToEntity(existingComment *entities.Comment) *entities.Comment {
	// Note: Comment entity doesn't have a Content field in the current structure
	// This might need to be updated based on the actual Comment entity design
	return existingComment
}

// CommentQueryParams represents query parameters for retrieving and filtering comments.
// This struct supports various filtering options including post association, user ownership,
// parent-child relationships, and status filtering, with pagination support.
type CommentQueryParams struct {
	// PostID filters comments by the post they belong to (optional)
	PostID *uuid.UUID `query:"post_id"`
	// UserID filters comments by the user who created them (optional)
	UserID *uuid.UUID `query:"user_id"`
	// ParentID filters comments by their parent comment for threaded discussions (optional)
	ParentID *uuid.UUID `query:"parent_id"`
	// Status filters comments by their moderation status (optional, defaults to "approved")
	Status string `query:"status"`
	// Limit controls the maximum number of comments returned (optional, 1-100, defaults to 10)
	Limit int `query:"limit" validate:"min=1,max=100"`
	// Offset controls the number of comments to skip for pagination (optional, minimum 0, defaults to 0)
	Offset int `query:"offset" validate:"min=0"`
}

// SetDefaults sets sensible default values for query parameters to ensure
// consistent behavior when optional parameters are not provided.
//
// Default Values:
// - Limit: 10 (reasonable page size for comment lists)
// - Offset: 0 (start from the beginning)
// - Status: "approved" (show only moderated comments by default)
func (q *CommentQueryParams) SetDefaults() {
	// Set default pagination limit if not specified or invalid
	if q.Limit <= 0 {
		q.Limit = 10
	}
	// Set default offset if not specified or negative
	if q.Offset < 0 {
		q.Offset = 0
	}
	// Set default status to approved for security (only show moderated content)
	if q.Status == "" {
		q.Status = "approved" // default to approved comments
	}
}
