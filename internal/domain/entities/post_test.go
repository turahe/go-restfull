package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPost_Success(t *testing.T) {
	title := "Test Post Title"
	content := "Test post content"
	slug := "test-post-title"
	status := "draft"
	authorID := uuid.New()

	post, err := entities.NewPost(title, content, slug, status, authorID)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, title, post.Title)
	assert.Equal(t, content, post.Content)
	assert.Equal(t, slug, post.Slug)
	assert.Equal(t, status, post.Status)
	assert.Equal(t, authorID, post.AuthorID)
	assert.NotEqual(t, uuid.Nil, post.ID)
	assert.False(t, post.CreatedAt.IsZero())
	assert.False(t, post.UpdatedAt.IsZero())
	assert.Nil(t, post.DeletedAt)
	assert.Nil(t, post.PublishedAt)
}

func TestNewPost_EmptyTitle(t *testing.T) {
	content := "Test post content"
	slug := "test-post-title"
	status := "draft"
	authorID := uuid.New()

	post, err := entities.NewPost("", content, slug, status, authorID)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "title is required", err.Error())
}

func TestNewPost_EmptyContent(t *testing.T) {
	title := "Test Post Title"
	slug := "test-post-title"
	status := "draft"
	authorID := uuid.New()

	post, err := entities.NewPost(title, "", slug, status, authorID)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "content is required", err.Error())
}

func TestNewPost_EmptySlug(t *testing.T) {
	title := "Test Post Title"
	content := "Test post content"
	status := "draft"
	authorID := uuid.New()

	post, err := entities.NewPost(title, content, "", status, authorID)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "slug is required", err.Error())
}

func TestNewPost_EmptyStatus(t *testing.T) {
	title := "Test Post Title"
	content := "Test post content"
	slug := "test-post-title"
	authorID := uuid.New()

	post, err := entities.NewPost(title, content, slug, "", authorID)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "status is required", err.Error())
}

func TestNewPost_EmptyAuthorID(t *testing.T) {
	title := "Test Post Title"
	content := "Test post content"
	slug := "test-post-title"
	status := "draft"

	post, err := entities.NewPost(title, content, slug, status, uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "author_id is required", err.Error())
}

func TestPost_UpdatePost(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "Old content", "old-slug", "draft", uuid.New())
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("New Title", "New content", "new-slug", "published")

	assert.NoError(t, err)
	assert.Equal(t, "New Title", post.Title)
	assert.Equal(t, "New content", post.Content)
	assert.Equal(t, "new-slug", post.Slug)
	assert.Equal(t, "published", post.Status)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_UpdatePost_PartialUpdate(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "Old content", "old-slug", "draft", uuid.New())
	originalTitle := post.Title
	originalSlug := post.Slug

	err := post.UpdatePost("", "New content", "", "published")

	assert.NoError(t, err)
	assert.Equal(t, originalTitle, post.Title) // Should remain unchanged
	assert.Equal(t, "New content", post.Content)
	assert.Equal(t, originalSlug, post.Slug) // Should remain unchanged
	assert.Equal(t, "published", post.Status)
}

func TestPost_UpdatePost_EmptyStrings(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "Old content", "old-slug", "draft", uuid.New())
	originalTitle := post.Title
	originalContent := post.Content
	originalSlug := post.Slug
	originalStatus := post.Status

	err := post.UpdatePost("", "", "", "")

	assert.NoError(t, err)
	assert.Equal(t, originalTitle, post.Title)
	assert.Equal(t, originalContent, post.Content)
	assert.Equal(t, originalSlug, post.Slug)
	assert.Equal(t, originalStatus, post.Status)
}

func TestPost_Publish(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "draft", uuid.New())
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.Publish()

	assert.Equal(t, "published", post.Status)
	assert.NotNil(t, post.PublishedAt)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, post.IsPublished())
}

func TestPost_Unpublish(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "published", uuid.New())
	post.PublishedAt = &time.Time{}
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.Unpublish()

	assert.Equal(t, "draft", post.Status)
	assert.Nil(t, post.PublishedAt)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, post.IsDraft())
}

func TestPost_SoftDelete(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "draft", uuid.New())
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.SoftDelete()

	assert.NotNil(t, post.DeletedAt)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, post.IsDeleted())
}

func TestPost_IsDeleted(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "draft", uuid.New())

	// Initially not deleted
	assert.False(t, post.IsDeleted())

	// After soft delete
	post.SoftDelete()
	assert.True(t, post.IsDeleted())
}

func TestPost_IsPublished(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "draft", uuid.New())

	// Initially not published
	assert.False(t, post.IsPublished())

	// After publishing
	post.Publish()
	assert.True(t, post.IsPublished())
}

func TestPost_IsPublished_WithoutPublishMethod(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "published", uuid.New())

	// Status is published but no PublishedAt timestamp
	assert.False(t, post.IsPublished())

	// Set PublishedAt timestamp
	now := time.Now()
	post.PublishedAt = &now
	assert.True(t, post.IsPublished())
}

func TestPost_IsDraft(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "draft", uuid.New())

	// Initially draft
	assert.True(t, post.IsDraft())

	// After publishing
	post.Publish()
	assert.False(t, post.IsDraft())
}

func TestPost_StatusTransitions(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "draft", uuid.New())

	// Initially draft
	assert.True(t, post.IsDraft())
	assert.False(t, post.IsPublished())

	// Publish
	post.Publish()
	assert.False(t, post.IsDraft())
	assert.True(t, post.IsPublished())

	// Unpublish
	post.Unpublish()
	assert.True(t, post.IsDraft())
	assert.False(t, post.IsPublished())
}

func TestPost_UpdatePost_OnlyTitle(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "Old content", "old-slug", "draft", uuid.New())
	originalContent := post.Content
	originalSlug := post.Slug
	originalStatus := post.Status

	err := post.UpdatePost("New Title", "", "", "")

	assert.NoError(t, err)
	assert.Equal(t, "New Title", post.Title)
	assert.Equal(t, originalContent, post.Content) // Should remain unchanged
	assert.Equal(t, originalSlug, post.Slug)       // Should remain unchanged
	assert.Equal(t, originalStatus, post.Status)   // Should remain unchanged
}

func TestPost_UpdatePost_OnlyContent(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "Old content", "old-slug", "draft", uuid.New())
	originalTitle := post.Title
	originalSlug := post.Slug
	originalStatus := post.Status

	err := post.UpdatePost("", "New content", "", "")

	assert.NoError(t, err)
	assert.Equal(t, originalTitle, post.Title) // Should remain unchanged
	assert.Equal(t, "New content", post.Content)
	assert.Equal(t, originalSlug, post.Slug)     // Should remain unchanged
	assert.Equal(t, originalStatus, post.Status) // Should remain unchanged
}

func TestPost_UpdatePost_OnlySlug(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "Old content", "old-slug", "draft", uuid.New())
	originalTitle := post.Title
	originalContent := post.Content
	originalStatus := post.Status

	err := post.UpdatePost("", "", "new-slug", "")

	assert.NoError(t, err)
	assert.Equal(t, originalTitle, post.Title)     // Should remain unchanged
	assert.Equal(t, originalContent, post.Content) // Should remain unchanged
	assert.Equal(t, "new-slug", post.Slug)
	assert.Equal(t, originalStatus, post.Status) // Should remain unchanged
}

func TestPost_UpdatePost_OnlyStatus(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "Old content", "old-slug", "draft", uuid.New())
	originalTitle := post.Title
	originalContent := post.Content
	originalSlug := post.Slug

	err := post.UpdatePost("", "", "", "published")

	assert.NoError(t, err)
	assert.Equal(t, originalTitle, post.Title)     // Should remain unchanged
	assert.Equal(t, originalContent, post.Content) // Should remain unchanged
	assert.Equal(t, originalSlug, post.Slug)       // Should remain unchanged
	assert.Equal(t, "published", post.Status)
}

func TestPost_SoftDelete_MultipleCalls(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "draft", uuid.New())

	// First soft delete
	post.SoftDelete()
	firstDeletedAt := post.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	post.SoftDelete()
	secondDeletedAt := post.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, post.IsDeleted())
}

func TestPost_Publish_AlreadyPublished(t *testing.T) {
	post, _ := entities.NewPost("Test Title", "Test content", "test-slug", "published", uuid.New())
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.Publish()

	// Should update timestamps even if already published
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, post.IsPublished())
}
