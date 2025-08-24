package requests

import (
	"github.com/google/uuid"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// CreateCommentRequest represents the request for creating a new comment
type CreateCommentRequest struct {
	Content  string     `json:"content" validate:"required,min=1,max=1000"`
	PostID   uuid.UUID  `json:"post_id" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// ToEntity transforms CreateCommentRequest to a Comment entity
func (r *CreateCommentRequest) ToEntity(authorID uuid.UUID) *entities.Comment {
	return &entities.Comment{
		ID:        uuid.New(),
		ModelID:   r.PostID,
		ModelType: "post",
		ParentID:  r.ParentID,
		CreatedBy: authorID,
		Status:    entities.CommentStatusPending,
	}
}

// UpdateCommentRequest represents the request for updating a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// ToEntity transforms UpdateCommentRequest to update an existing Comment entity
func (r *UpdateCommentRequest) ToEntity(existingComment *entities.Comment) *entities.Comment {
	// Note: Comment entity doesn't have a Content field in the current structure
	// This might need to be updated based on the actual Comment entity design
	return existingComment
}

// CommentQueryParams represents query parameters for comment listing
type CommentQueryParams struct {
	PostID   *uuid.UUID `query:"post_id"`
	UserID   *uuid.UUID `query:"user_id"`
	ParentID *uuid.UUID `query:"parent_id"`
	Status   string     `query:"status"`
	Limit    int        `query:"limit" validate:"min=1,max=100"`
	Offset   int        `query:"offset" validate:"min=0"`
}

// SetDefaults sets default values for query parameters
func (q *CommentQueryParams) SetDefaults() {
	if q.Limit <= 0 {
		q.Limit = 10
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	if q.Status == "" {
		q.Status = "approved" // default to approved comments
	}
}
