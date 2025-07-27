package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Comment represents the core comment domain entity
type Comment struct {
	ID        uuid.UUID  `json:"id"`
	Content   string     `json:"content"`
	PostID    uuid.UUID  `json:"post_id"`
	UserID    uuid.UUID  `json:"user_id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewComment creates a new comment with validation
func NewComment(content string, postID, userID uuid.UUID, parentID *uuid.UUID, status string) (*Comment, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}
	if postID == uuid.Nil {
		return nil, errors.New("post_id is required")
	}
	if userID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}
	if status == "" {
		status = "pending" // default status
	}

	now := time.Now()
	return &Comment{
		ID:        uuid.New(),
		Content:   content,
		PostID:    postID,
		UserID:    userID,
		ParentID:  parentID,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateComment updates comment information
func (c *Comment) UpdateComment(content, status string) error {
	if content != "" {
		c.Content = content
	}
	if status != "" {
		c.Status = status
	}
	c.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the comment as deleted
func (c *Comment) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
}

// IsDeleted checks if the comment is deleted
func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsApproved checks if the comment is approved
func (c *Comment) IsApproved() bool {
	return c.Status == "approved"
}

// IsPending checks if the comment is pending
func (c *Comment) IsPending() bool {
	return c.Status == "pending"
}

// IsRejected checks if the comment is rejected
func (c *Comment) IsRejected() bool {
	return c.Status == "rejected"
}

// Approve approves the comment
func (c *Comment) Approve() {
	c.Status = "approved"
	c.UpdatedAt = time.Now()
}

// Reject rejects the comment
func (c *Comment) Reject() {
	c.Status = "rejected"
	c.UpdatedAt = time.Now()
}

// IsReply checks if the comment is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}
