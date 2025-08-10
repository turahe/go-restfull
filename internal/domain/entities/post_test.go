package entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/turahe/go-restfull/internal/domain/entities"
)

func TestNewPost_Success(t *testing.T) {
	title := "Test Post Title"
	slug := "test-post-title"
	subtitle := "Test Post Subtitle"
	description := "Test post description"
	language := "en"
	layout := "default"
	isSticky := false
	var publishedAt *time.Time = nil

	post, err := entities.NewPost(title, slug, subtitle, description, language, layout, isSticky, publishedAt)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, title, post.Title)
	assert.Equal(t, slug, post.Slug)
	assert.Equal(t, subtitle, post.Subtitle)
	assert.Equal(t, description, post.Description)
	assert.Equal(t, language, post.Language)
	assert.Equal(t, layout, post.Layout)
	assert.Equal(t, isSticky, post.IsSticky)
	assert.NotEqual(t, uuid.Nil, post.ID)
	assert.False(t, post.CreatedAt.IsZero())
	assert.False(t, post.UpdatedAt.IsZero())
	assert.Nil(t, post.DeletedAt)
	assert.Nil(t, post.PublishedAt)
}

func TestNewPost_EmptyTitle(t *testing.T) {
	slug := "test-post-title"
	subtitle := "Test Post Subtitle"
	description := "Test post description"
	language := "en"
	layout := "default"
	isSticky := false
	var publishedAt *time.Time = nil

	post, err := entities.NewPost("", slug, subtitle, description, language, layout, isSticky, publishedAt)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "title is required", err.Error())
}

func TestNewPost_EmptySlug(t *testing.T) {
	title := "Test Post Title"
	subtitle := "Test Post Subtitle"
	description := "Test post description"
	language := "en"
	layout := "default"
	isSticky := false
	var publishedAt *time.Time = nil

	post, err := entities.NewPost(title, "", subtitle, description, language, layout, isSticky, publishedAt)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "slug is required", err.Error())
}

func TestNewPost_EmptySubtitle(t *testing.T) {
	title := "Test Post Title"
	slug := "test-post-title"
	description := "Test post description"
	language := "en"
	layout := "default"
	isSticky := false
	var publishedAt *time.Time = nil

	post, err := entities.NewPost(title, slug, "", description, language, layout, isSticky, publishedAt)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "subtitle is required", err.Error())
}

func TestNewPost_EmptyDescription(t *testing.T) {
	title := "Test Post Title"
	slug := "test-post-title"
	subtitle := "Test Post Subtitle"
	language := "en"
	layout := "default"
	isSticky := false
	var publishedAt *time.Time = nil

	post, err := entities.NewPost(title, slug, subtitle, "", language, layout, isSticky, publishedAt)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "description is required", err.Error())
}

func TestNewPost_EmptyLanguage(t *testing.T) {
	title := "Test Post Title"
	slug := "test-post-title"
	subtitle := "Test Post Subtitle"
	description := "Test post description"
	layout := "default"
	isSticky := false
	var publishedAt *time.Time = nil

	post, err := entities.NewPost(title, slug, subtitle, description, "", layout, isSticky, publishedAt)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "language is required", err.Error())
}

func TestNewPost_EmptyLayout(t *testing.T) {
	title := "Test Post Title"
	slug := "test-post-title"
	subtitle := "Test Post Subtitle"
	description := "Test post description"
	language := "en"
	isSticky := false
	var publishedAt *time.Time = nil

	post, err := entities.NewPost(title, slug, subtitle, description, language, "", isSticky, publishedAt)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "layout is required", err.Error())
}

func TestPost_UpdatePost(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("New Title", "new-slug", "New Subtitle", "New description", "en", "new-layout", true, nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Title", post.Title)
	assert.Equal(t, "new-slug", post.Slug)
	assert.Equal(t, "New Subtitle", post.Subtitle)
	assert.Equal(t, "New description", post.Description)
	assert.Equal(t, "en", post.Language)
	assert.Equal(t, "new-layout", post.Layout)
	assert.True(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_UpdatePost_PartialUpdate(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("New Title", "", "", "", "", "", false, nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Title", post.Title)
	assert.Equal(t, "old-slug", post.Slug)
	assert.Equal(t, "Old Subtitle", post.Subtitle)
	assert.Equal(t, "Old description", post.Description)
	assert.Equal(t, "en", post.Language)
	assert.Equal(t, "old-layout", post.Layout)
	assert.False(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_UpdatePost_EmptyStrings(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("", "", "", "", "", "", false, nil)

	assert.NoError(t, err)
	assert.Equal(t, "Old Title", post.Title)
	assert.Equal(t, "old-slug", post.Slug)
	assert.Equal(t, "Old Subtitle", post.Subtitle)
	assert.Equal(t, "Old description", post.Description)
	assert.Equal(t, "en", post.Language)
	assert.Equal(t, "old-layout", post.Layout)
	assert.False(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_Publish(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.Publish()

	assert.NotNil(t, post.PublishedAt)
	assert.True(t, post.PublishedAt.After(originalUpdatedAt))
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, post.IsPublished())
}

func TestPost_Unpublish(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)
	post.Publish()
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.Unpublish()

	assert.Nil(t, post.PublishedAt)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, post.IsPublished())
	assert.True(t, post.IsDraft())
}

func TestPost_SoftDelete(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.SoftDelete()

	assert.NotNil(t, post.DeletedAt)
	assert.True(t, post.DeletedAt.After(originalUpdatedAt))
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, post.IsDeleted())
}

func TestPost_IsDeleted(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)

	assert.False(t, post.IsDeleted())

	post.SoftDelete()

	assert.True(t, post.IsDeleted())
}

func TestPost_IsPublished(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)

	assert.False(t, post.IsPublished())

	post.Publish()

	assert.True(t, post.IsPublished())
}

func TestPost_IsPublished_WithoutPublishMethod(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)

	assert.False(t, post.IsPublished())
}

func TestPost_IsDraft(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)

	assert.True(t, post.IsDraft())

	post.Publish()

	assert.False(t, post.IsDraft())
}

func TestPost_StatusTransitions(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)

	// Initially draft
	assert.True(t, post.IsDraft())
	assert.False(t, post.IsPublished())

	// Publish
	post.Publish()
	assert.True(t, post.IsPublished())
	assert.False(t, post.IsDraft())

	// Unpublish
	post.Unpublish()
	assert.False(t, post.IsPublished())
	assert.True(t, post.IsDraft())
}

func TestPost_UpdatePost_OnlyTitle(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("New Title", "", "", "", "", "", false, nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Title", post.Title)
	assert.Equal(t, "old-slug", post.Slug)
	assert.Equal(t, "Old Subtitle", post.Subtitle)
	assert.Equal(t, "Old description", post.Description)
	assert.Equal(t, "en", post.Language)
	assert.Equal(t, "old-layout", post.Layout)
	assert.False(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_UpdatePost_OnlySubtitle(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("", "", "New Subtitle", "", "", "", false, nil)

	assert.NoError(t, err)
	assert.Equal(t, "Old Title", post.Title)
	assert.Equal(t, "old-slug", post.Slug)
	assert.Equal(t, "New Subtitle", post.Subtitle)
	assert.Equal(t, "Old description", post.Description)
	assert.Equal(t, "en", post.Language)
	assert.Equal(t, "old-layout", post.Layout)
	assert.False(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_UpdatePost_OnlyDescription(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("", "", "", "New description", "", "", false, nil)

	assert.NoError(t, err)
	assert.Equal(t, "Old Title", post.Title)
	assert.Equal(t, "old-slug", post.Slug)
	assert.Equal(t, "Old Subtitle", post.Subtitle)
	assert.Equal(t, "New description", post.Description)
	assert.Equal(t, "en", post.Language)
	assert.Equal(t, "old-layout", post.Layout)
	assert.False(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_UpdatePost_OnlyLanguage(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("", "", "", "", "fr", "", false, nil)

	assert.NoError(t, err)
	assert.Equal(t, "Old Title", post.Title)
	assert.Equal(t, "old-slug", post.Slug)
	assert.Equal(t, "Old Subtitle", post.Subtitle)
	assert.Equal(t, "Old description", post.Description)
	assert.Equal(t, "fr", post.Language)
	assert.Equal(t, "old-layout", post.Layout)
	assert.False(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_UpdatePost_OnlyLayout(t *testing.T) {
	post, _ := entities.NewPost("Old Title", "old-slug", "Old Subtitle", "Old description", "en", "old-layout", false, nil)
	originalUpdatedAt := post.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := post.UpdatePost("", "", "", "", "", "new-layout", false, nil)

	assert.NoError(t, err)
	assert.Equal(t, "Old Title", post.Title)
	assert.Equal(t, "old-slug", post.Slug)
	assert.Equal(t, "Old Subtitle", post.Subtitle)
	assert.Equal(t, "Old description", post.Description)
	assert.Equal(t, "en", post.Language)
	assert.Equal(t, "new-layout", post.Layout)
	assert.False(t, post.IsSticky)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPost_SoftDelete_MultipleCalls(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)
	originalDeletedAt := post.DeletedAt

	post.SoftDelete()
	firstDeletedAt := post.DeletedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	post.SoftDelete()
	secondDeletedAt := post.DeletedAt

	assert.NotEqual(t, originalDeletedAt, firstDeletedAt)
	assert.NotEqual(t, originalDeletedAt, secondDeletedAt)
	assert.True(t, post.IsDeleted())
}

func TestPost_Publish_AlreadyPublished(t *testing.T) {
	post, _ := entities.NewPost("Test Post", "test-post", "Test Subtitle", "Test description", "en", "default", false, nil)
	post.Publish()
	originalPublishedAt := post.PublishedAt

	post.Publish()

	// Verify that PublishedAt didn't change
	assert.Equal(t, originalPublishedAt, post.PublishedAt)
	// Verify that the post is still published
	assert.True(t, post.IsPublished())
	// Note: UpdatedAt may or may not change depending on timing, but that's not the main test concern
	// The main test is that calling Publish on an already published post doesn't change the PublishedAt
}
