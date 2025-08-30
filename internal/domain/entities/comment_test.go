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
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, modelType, comment.ModelType)
	assert.Equal(t, modelID, comment.ModelID)
	assert.Equal(t, content, comment.Content)
	assert.Equal(t, &parentID, comment.ParentID)
	assert.Equal(t, entities.CommentStatusPending, comment.Status)
	assert.Equal(t, createdBy, comment.CreatedBy)
	assert.Equal(t, createdBy, comment.UpdatedBy)
	assert.NotEqual(t, uuid.Nil, comment.ID)
	assert.False(t, comment.CreatedAt.IsZero())
	assert.False(t, comment.UpdatedAt.IsZero())
	assert.Nil(t, comment.DeletedAt)
}

func TestNewComment_WithoutParentID(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	createdBy := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, content, nil, entities.CommentStatusApproved, createdBy)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, modelType, comment.ModelType)
	assert.Equal(t, modelID, comment.ModelID)
	assert.Equal(t, content, comment.Content)
	assert.Nil(t, comment.ParentID)
	assert.Equal(t, entities.CommentStatusApproved, comment.Status)
	assert.Equal(t, createdBy, comment.CreatedBy)
}

func TestNewComment_DefaultStatus(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, content, &parentID, "pending", createdBy)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, entities.CommentStatusPending, comment.Status)
}

func TestNewComment_EmptyModelType(t *testing.T) {
	modelType := ""
	modelID := uuid.New()
	content := "Great post!"
	createdBy := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, content, nil, entities.CommentStatusPending, createdBy)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "model_type is required", err.Error())
}

func TestNewComment_EmptyModelID(t *testing.T) {
	modelType := "post"
	modelID := uuid.Nil
	content := "Great post!"
	createdBy := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, content, nil, entities.CommentStatusPending, createdBy)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "model_id is required", err.Error())
}

func TestNewComment_EmptyContent(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := ""
	createdBy := uuid.New()

	comment, err := entities.NewComment(modelType, modelID, content, nil, entities.CommentStatusPending, createdBy)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "content is required", err.Error())
}

func TestNewComment_EmptyCreatedBy(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	createdBy := uuid.Nil

	comment, err := entities.NewComment(modelType, modelID, content, nil, entities.CommentStatusPending, createdBy)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "created_by is required", err.Error())
}

func TestComment_UpdateComment(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)
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
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	err := comment.UpdateComment(entities.CommentStatusApproved)

	assert.NoError(t, err)
	assert.Equal(t, entities.CommentStatusApproved, comment.Status)
}

func TestComment_UpdateComment_OnlyContent(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)
	originalStatus := comment.Status

	err := comment.UpdateComment(entities.CommentStatusApproved)

	assert.NoError(t, err)
	assert.NotEqual(t, originalStatus, comment.Status) // Status should change
	assert.Equal(t, entities.CommentStatusApproved, comment.Status)
}

func TestComment_SoftDelete(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)
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
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	// Initially not deleted
	assert.False(t, comment.IsDeleted())

	// After soft delete
	comment.SoftDelete()
	assert.True(t, comment.IsDeleted())
}

func TestComment_IsApproved(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	// Initially not approved
	assert.False(t, comment.IsApproved())

	// After approval
	comment.Status = entities.CommentStatusApproved
	assert.True(t, comment.IsApproved())
}

func TestComment_IsPending(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	// Initially pending
	assert.True(t, comment.IsPending())

	// After status change
	comment.Status = entities.CommentStatusApproved
	assert.False(t, comment.IsPending())
}

func TestComment_IsRejected(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	// Initially not rejected
	assert.False(t, comment.IsRejected())

	// After rejection
	comment.Status = "rejected"
	assert.True(t, comment.IsRejected())
}

func TestComment_Approve(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)
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
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)
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
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	// Should be a reply since parent ID is set
	assert.True(t, comment.IsReply())

	// Remove parent ID to make it not a reply
	comment.ParentID = nil
	assert.False(t, comment.IsReply())
}

func TestComment_IsReply_WithParentID(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

	// Should be a reply since parent ID is set
	assert.True(t, comment.IsReply())
}

func TestComment_StatusTransitions(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

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
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)
	originalStatus := comment.Status

	err := comment.UpdateComment("")

	assert.NoError(t, err)
	assert.Equal(t, originalStatus, comment.Status) // Status should remain unchanged when empty string is passed
}

func TestComment_SoftDelete_MultipleCalls(t *testing.T) {
	modelType := "post"
	modelID := uuid.New()
	content := "Great post!"
	parentID := uuid.New()
	createdBy := uuid.New()

	comment, _ := entities.NewComment(modelType, modelID, content, &parentID, entities.CommentStatusPending, createdBy)

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
