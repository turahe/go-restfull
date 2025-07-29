package requests

import (
	"github.com/google/uuid"
)

// CreateCommentRequest represents the request for creating a new comment
type CreateCommentRequest struct {
	Content  string     `json:"content" validate:"required,min=1,max=1000"`
	PostID   uuid.UUID  `json:"post_id" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// UpdateCommentRequest represents the request for updating a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
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
