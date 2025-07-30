package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewComment_Success(t *testing.T) {
	content := "Test comment content"
	postID := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()

	comment, err := entities.NewComment(content, postID, userID, &parentID, "pending")

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, content, comment.Content)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, userID, comment.UserID)
	assert.Equal(t, &parentID, comment.ParentID)
	assert.Equal(t, "pending", comment.Status)
	assert.NotEqual(t, uuid.Nil, comment.ID)
	assert.False(t, comment.CreatedAt.IsZero())
	assert.False(t, comment.UpdatedAt.IsZero())
	assert.Nil(t, comment.DeletedAt)
}

func TestNewComment_WithoutParentID(t *testing.T) {
	content := "Test comment content"
	postID := uuid.New()
	userID := uuid.New()

	comment, err := entities.NewComment(content, postID, userID, nil, "approved")

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, content, comment.Content)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, userID, comment.UserID)
	assert.Nil(t, comment.ParentID)
	assert.Equal(t, "approved", comment.Status)
}

func TestNewComment_DefaultStatus(t *testing.T) {
	content := "Test comment content"
	postID := uuid.New()
	userID := uuid.New()

	comment, err := entities.NewComment(content, postID, userID, nil, "")

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "pending", comment.Status)
}

func TestNewComment_EmptyContent(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	comment, err := entities.NewComment("", postID, userID, nil, "pending")

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "content is required", err.Error())
}

func TestNewComment_EmptyPostID(t *testing.T) {
	content := "Test comment content"
	userID := uuid.New()

	comment, err := entities.NewComment(content, uuid.Nil, userID, nil, "pending")

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "post_id is required", err.Error())
}

func TestNewComment_EmptyUserID(t *testing.T) {
	content := "Test comment content"
	postID := uuid.New()

	comment, err := entities.NewComment(content, postID, uuid.Nil, nil, "pending")

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "user_id is required", err.Error())
}

func TestComment_UpdateComment(t *testing.T) {
	comment, _ := entities.NewComment("Original content", uuid.New(), uuid.New(), nil, "pending")
	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := comment.UpdateComment("Updated content", "approved")

	assert.NoError(t, err)
	assert.Equal(t, "Updated content", comment.Content)
	assert.Equal(t, "approved", comment.Status)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
}

func TestComment_UpdateComment_PartialUpdate(t *testing.T) {
	comment, _ := entities.NewComment("Original content", uuid.New(), uuid.New(), nil, "pending")
	originalContent := comment.Content

	err := comment.UpdateComment("", "approved")

	assert.NoError(t, err)
	assert.Equal(t, originalContent, comment.Content) // Should remain unchanged
	assert.Equal(t, "approved", comment.Status)
}

func TestComment_UpdateComment_OnlyContent(t *testing.T) {
	comment, _ := entities.NewComment("Original content", uuid.New(), uuid.New(), nil, "pending")
	originalStatus := comment.Status

	err := comment.UpdateComment("Updated content", "")

	assert.NoError(t, err)
	assert.Equal(t, "Updated content", comment.Content)
	assert.Equal(t, originalStatus, comment.Status) // Should remain unchanged
}

func TestComment_SoftDelete(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")
	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	comment.SoftDelete()

	assert.NotNil(t, comment.DeletedAt)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, comment.IsDeleted())
}

func TestComment_IsDeleted(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")

	// Initially not deleted
	assert.False(t, comment.IsDeleted())

	// After soft delete
	comment.SoftDelete()
	assert.True(t, comment.IsDeleted())
}

func TestComment_IsApproved(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")

	// Initially not approved
	assert.False(t, comment.IsApproved())

	// After approval
	comment.Status = "approved"
	assert.True(t, comment.IsApproved())
}

func TestComment_IsPending(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")

	// Initially pending
	assert.True(t, comment.IsPending())

	// After status change
	comment.Status = "approved"
	assert.False(t, comment.IsPending())
}

func TestComment_IsRejected(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")

	// Initially not rejected
	assert.False(t, comment.IsRejected())

	// After rejection
	comment.Status = "rejected"
	assert.True(t, comment.IsRejected())
}

func TestComment_Approve(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")
	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	comment.Approve()

	assert.Equal(t, "approved", comment.Status)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, comment.IsApproved())
}

func TestComment_Reject(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")
	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	comment.Reject()

	assert.Equal(t, "rejected", comment.Status)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, comment.IsRejected())
}

func TestComment_IsReply(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")

	// Initially not a reply
	assert.False(t, comment.IsReply())

	// After setting parent ID
	parentID := uuid.New()
	comment.ParentID = &parentID
	assert.True(t, comment.IsReply())
}

func TestComment_IsReply_WithParentID(t *testing.T) {
	parentID := uuid.New()
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), &parentID, "pending")

	// Should be a reply since parent ID is set
	assert.True(t, comment.IsReply())
}

func TestComment_StatusTransitions(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")

	// Test status transitions
	assert.True(t, comment.IsPending())
	assert.False(t, comment.IsApproved())
	assert.False(t, comment.IsRejected())

	comment.Approve()
	assert.False(t, comment.IsPending())
	assert.True(t, comment.IsApproved())
	assert.False(t, comment.IsRejected())

	comment.Reject()
	assert.False(t, comment.IsPending())
	assert.False(t, comment.IsApproved())
	assert.True(t, comment.IsRejected())
}

func TestComment_UpdateComment_EmptyStrings(t *testing.T) {
	comment, _ := entities.NewComment("Original content", uuid.New(), uuid.New(), nil, "pending")
	originalContent := comment.Content
	originalStatus := comment.Status

	err := comment.UpdateComment("", "")

	assert.NoError(t, err)
	assert.Equal(t, originalContent, comment.Content)
	assert.Equal(t, originalStatus, comment.Status)
}

func TestComment_SoftDelete_MultipleCalls(t *testing.T) {
	comment, _ := entities.NewComment("Test content", uuid.New(), uuid.New(), nil, "pending")

	// First soft delete
	comment.SoftDelete()
	firstDeletedAt := comment.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	comment.SoftDelete()
	secondDeletedAt := comment.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, comment.IsDeleted())
}
