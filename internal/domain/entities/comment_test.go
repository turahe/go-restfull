package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewComment_Success(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, modelType, comment.ModelType)
	assert.Equal(t, modelID, comment.ModelID)
	assert.Equal(t, &parentID, comment.ParentID)
	assert.Equal(t, entities.CommentStatusPending, comment.Status)
	assert.NotEqual(t, uuid.Nil, comment.ID)
	assert.False(t, comment.CreatedAt.IsZero())
	assert.False(t, comment.UpdatedAt.IsZero())
	assert.Nil(t, comment.DeletedAt)
}

func TestNewComment_WithoutParentID(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, nil, entities.CommentStatusApproved)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, modelType, comment.ModelType)
	assert.Equal(t, modelID, comment.ModelID)
	assert.Nil(t, comment.ParentID)
	assert.Equal(t, entities.CommentStatusApproved, comment.Status)
}

func TestNewComment_DefaultStatus(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, &parentID, "pending")

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, entities.CommentStatusPending, comment.Status)
}

func TestNewComment_EmptyContent(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, nil, entities.CommentStatusPending)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "model_type is required", err.Error())
}

func TestNewComment_EmptyPostID(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, nil, entities.CommentStatusPending)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "model_id is required", err.Error())
}

func TestNewComment_EmptyUserID(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, nil, entities.CommentStatusPending)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "model_id is required", err.Error())
}

func TestComment_UpdateComment(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)
	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := comment.UpdateComment(entities.CommentStatusApproved)

	assert.NoError(t, err)
	assert.Equal(t, entities.CommentStatusApproved, comment.Status)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
}

func TestComment_UpdateComment_PartialUpdate(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	err := comment.UpdateComment(entities.CommentStatusApproved)

	assert.NoError(t, err)
	assert.Equal(t, entities.CommentStatusApproved, comment.Status)
}

func TestComment_UpdateComment_OnlyContent(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)
	originalStatus := comment.Status

	err := comment.UpdateComment(entities.CommentStatusApproved)

	assert.NoError(t, err)
	assert.Equal(t, originalStatus, comment.Status) // Should remain unchanged
}

func TestComment_SoftDelete(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)
	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	comment.SoftDelete()

	assert.NotNil(t, comment.DeletedAt)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, comment.IsDeleted())
}

func TestComment_IsDeleted(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	// Initially not deleted
	assert.False(t, comment.IsDeleted())

	// After soft delete
	comment.SoftDelete()
	assert.True(t, comment.IsDeleted())
}

func TestComment_IsApproved(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	// Initially not approved
	assert.False(t, comment.IsApproved())

	// After approval
	comment.Status = entities.CommentStatusApproved
	assert.True(t, comment.IsApproved())
}

func TestComment_IsPending(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	// Initially pending
	assert.True(t, comment.IsPending())

	// After status change
	comment.Status = entities.CommentStatusApproved
	assert.False(t, comment.IsPending())
}

func TestComment_IsRejected(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	// Initially not rejected
	assert.False(t, comment.IsRejected())

	// After rejection
	comment.Status = "rejected"
	assert.True(t, comment.IsRejected())
}

func TestComment_Approve(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	comment.Approve()

	assert.Equal(t, entities.CommentStatusApproved, comment.Status)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, comment.IsApproved())
}

func TestComment_Reject(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	originalUpdatedAt := comment.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	comment.Reject()

	assert.Equal(t, entities.CommentStatusRejected, comment.Status)
	assert.True(t, comment.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, comment.IsRejected())
}

func TestComment_IsReply(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	// Initially not a reply
	assert.False(t, comment.IsReply())

	// After setting parent ID
	comment.ParentID = &parentID
	assert.True(t, comment.IsReply())
}

func TestComment_IsReply_WithParentID(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()
	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

	// Should be a reply since parent ID is set
	assert.True(t, comment.IsReply())
}

func TestComment_StatusTransitions(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

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
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)
	originalStatus := comment.Status

	err := comment.UpdateComment(entities.CommentStatusApproved)

	assert.NoError(t, err)
	assert.Equal(t, originalStatus, comment.Status)
}

func TestComment_SoftDelete_MultipleCalls(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	parentID := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, &parentID, entities.CommentStatusPending)

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
