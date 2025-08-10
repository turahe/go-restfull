// Package entities provides domain entities for the application.
// This package contains the core business logic and data structures
// that represent the fundamental concepts of the system.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// CommentStatus represents the current state of a comment in the system.
// Comments go through a workflow: pending → approved/rejected
type CommentStatus string

const (
	// CommentStatusPending indicates a comment is waiting for moderation
	CommentStatusPending CommentStatus = "pending"
	// CommentStatusApproved indicates a comment has been approved and is visible
	CommentStatusApproved CommentStatus = "approved"
	// CommentStatusRejected indicates a comment has been rejected and is not visible
	CommentStatusRejected CommentStatus = "rejected"
)

// Comment represents a user-generated comment on any entity in the system.
// Comments support hierarchical threading (replies to comments) and can be
// associated with any model type (posts, articles, etc.).
//
// The entity uses a polymorphic relationship pattern where ModelID and ModelType
// determine what the comment is attached to.
type Comment struct {
	// ID is the unique identifier for the comment
	ID uuid.UUID `json:"id"`

	// ModelID is the ID of the entity this comment belongs to (e.g., post ID, article ID)
	ModelID uuid.UUID `json:"model_id"`

	// ModelType is the type of entity this comment belongs to (e.g., "post", "article")
	ModelType string `json:"model_type"`

	// ParentID is the ID of the parent comment if this is a reply, nil for top-level comments
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// Status indicates the current moderation state of the comment
	Status CommentStatus `json:"status"`

	// RecordLeft is used for nested set model implementation (left boundary)
	// This field enables efficient tree traversal and querying
	RecordLeft *uint64 `json:"record_left,omitempty"`

	// RecordRight is used for nested set model implementation (right boundary)
	// This field enables efficient tree traversal and querying
	RecordRight *uint64 `json:"record_right,omitempty"`

	// RecordOrdering determines the display order of comments at the same level
	RecordOrdering *uint64 `json:"record_ordering,omitempty"`

	// RecordDepth indicates how deeply nested this comment is in the thread
	RecordDepth *uint64 `json:"record_depth,omitempty"`

	// CreatedBy is the ID of the user who created the comment
	CreatedBy uuid.UUID `json:"created_by"`

	// UpdatedBy is the ID of the user who last updated the comment
	UpdatedBy uuid.UUID `json:"updated_by"`

	// DeletedBy is the ID of the user who deleted the comment (soft delete)
	DeletedBy *uuid.UUID `json:"deleted_by,omitempty"`

	// CreatedAt is the timestamp when the comment was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the comment was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// DeletedAt is the timestamp when the comment was soft deleted, nil if not deleted
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewComment creates a new comment with basic validation.
//
// Parameters:
//   - modelType: The type of entity being commented on (e.g., "post", "article")
//   - modelID: The ID of the entity being commented on
//   - parentID: Optional parent comment ID for replies, nil for top-level comments
//   - status: The initial status of the comment (defaults to "pending" if empty)
//
// Returns:
//   - A new Comment instance with generated ID and timestamps
//   - An error if validation fails
//
// Example:
//
//	comment, err := NewComment("post", postID, nil, CommentStatusPending)
//	if err != nil {
//	    // handle error
//	}
func NewComment(modelType string, modelID uuid.UUID, parentID *uuid.UUID, status CommentStatus) (*Comment, error) {
	// Validate required fields
	if modelType == "" {
		return nil, errors.New("model_type is required")
	}
	if modelID == uuid.Nil {
		return nil, errors.New("model_id is required")
	}

	// Set default status if none provided
	if status == "" {
		status = CommentStatusPending // default status
	}

	// Create comment with current timestamp
	now := time.Now()
	return &Comment{
		ID:        uuid.New(),
		ModelID:   modelID,
		ModelType: modelType,
		ParentID:  parentID,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateComment updates the comment's status and marks it as modified.
// This method is typically used by moderators to change comment status.
//
// Parameters:
//   - status: The new status to assign to the comment
//
// Returns:
//   - An error if the update fails
//
// Example:
//
//	err := comment.UpdateComment(CommentStatusApproved)
//	if err != nil {
//	    // handle error
//	}
func (c *Comment) UpdateComment(status CommentStatus) error {
	// Update status if provided
	if status != "" {
		c.Status = status
	}

	// Always update the modification timestamp
	c.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the comment as deleted without removing it from the database.
// This preserves the comment history while hiding it from normal queries.
// The comment can be restored by setting DeletedAt back to nil.
func (c *Comment) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
}

// IsDeleted checks if the comment has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsApproved checks if the comment has been approved by a moderator.
// Approved comments are visible to all users.
func (c *Comment) IsApproved() bool {
	return c.Status == CommentStatusApproved
}

// IsPending checks if the comment is waiting for moderation.
// Pending comments are typically only visible to the author and moderators.
func (c *Comment) IsPending() bool {
	return c.Status == CommentStatusPending
}

// IsRejected checks if the comment has been rejected by a moderator.
// Rejected comments are not visible to regular users.
func (c *Comment) IsRejected() bool {
	return c.Status == CommentStatusRejected
}

// Approve changes the comment status to approved and updates the modification timestamp.
// This method is typically called by moderators to approve pending comments.
func (c *Comment) Approve() {
	c.Status = CommentStatusApproved
	c.UpdatedAt = time.Now()
}

// Reject changes the comment status to rejected and updates the modification timestamp.
// This method is typically called by moderators to reject inappropriate comments.
func (c *Comment) Reject() {
	c.Status = CommentStatusRejected
	c.UpdatedAt = time.Now()
}

// IsReply checks if this comment is a reply to another comment.
// Returns true if ParentID is not nil, false for top-level comments.
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}
