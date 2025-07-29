package entities_test

import (
	"testing"
	"time"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewMedia_Success(t *testing.T) {
	fileName := "test-image.jpg"
	originalName := "original-image.jpg"
	mimeType := "image/jpeg"
	path := "/uploads/images/test-image.jpg"
	url := "https://example.com/uploads/images/test-image.jpg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, originalName, mimeType, path, url, size, userID)

	assert.NoError(t, err)
	assert.NotNil(t, media)
	assert.Equal(t, fileName, media.FileName)
	assert.Equal(t, originalName, media.OriginalName)
	assert.Equal(t, mimeType, media.MimeType)
	assert.Equal(t, path, media.Path)
	assert.Equal(t, url, media.URL)
	assert.Equal(t, size, media.Size)
	assert.Equal(t, userID, media.UserID)
	assert.NotEqual(t, uuid.Nil, media.ID)
	assert.False(t, media.CreatedAt.IsZero())
	assert.False(t, media.UpdatedAt.IsZero())
	assert.Nil(t, media.DeletedAt)
}

func TestNewMedia_EmptyFileName(t *testing.T) {
	originalName := "original-image.jpg"
	mimeType := "image/jpeg"
	path := "/uploads/images/test-image.jpg"
	url := "https://example.com/uploads/images/test-image.jpg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia("", originalName, mimeType, path, url, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "file_name is required", err.Error())
}

func TestNewMedia_EmptyOriginalName(t *testing.T) {
	fileName := "test-image.jpg"
	mimeType := "image/jpeg"
	path := "/uploads/images/test-image.jpg"
	url := "https://example.com/uploads/images/test-image.jpg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, "", mimeType, path, url, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "original_name is required", err.Error())
}

func TestNewMedia_EmptyMimeType(t *testing.T) {
	fileName := "test-image.jpg"
	originalName := "original-image.jpg"
	path := "/uploads/images/test-image.jpg"
	url := "https://example.com/uploads/images/test-image.jpg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, originalName, "", path, url, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "mime_type is required", err.Error())
}

func TestNewMedia_EmptyPath(t *testing.T) {
	fileName := "test-image.jpg"
	originalName := "original-image.jpg"
	mimeType := "image/jpeg"
	url := "https://example.com/uploads/images/test-image.jpg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, originalName, mimeType, "", url, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "path is required", err.Error())
}

func TestNewMedia_EmptyURL(t *testing.T) {
	fileName := "test-image.jpg"
	originalName := "original-image.jpg"
	mimeType := "image/jpeg"
	path := "/uploads/images/test-image.jpg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, originalName, mimeType, path, "", size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "url is required", err.Error())
}

func TestNewMedia_ZeroSize(t *testing.T) {
	fileName := "test-image.jpg"
	originalName := "original-image.jpg"
	mimeType := "image/jpeg"
	path := "/uploads/images/test-image.jpg"
	url := "https://example.com/uploads/images/test-image.jpg"
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, originalName, mimeType, path, url, 0, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "size must be greater than 0", err.Error())
}

func TestNewMedia_NegativeSize(t *testing.T) {
	fileName := "test-image.jpg"
	originalName := "original-image.jpg"
	mimeType := "image/jpeg"
	path := "/uploads/images/test-image.jpg"
	url := "https://example.com/uploads/images/test-image.jpg"
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, originalName, mimeType, path, url, -1024, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "size must be greater than 0", err.Error())
}

func TestNewMedia_EmptyUserID(t *testing.T) {
	fileName := "test-image.jpg"
	originalName := "original-image.jpg"
	mimeType := "image/jpeg"
	path := "/uploads/images/test-image.jpg"
	url := "https://example.com/uploads/images/test-image.jpg"
	size := int64(1024)

	media, err := entities.NewMedia(fileName, originalName, mimeType, path, url, size, uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "user_id is required", err.Error())
}

func TestMedia_UpdateMedia(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "image/jpeg", "/old/path", "https://old-url.com", 1024, uuid.New())
	originalUpdatedAt := media.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := media.UpdateMedia("new-file.jpg", "new-original.jpg", "image/png", "/new/path", "https://new-url.com", 2048)

	assert.NoError(t, err)
	assert.Equal(t, "new-file.jpg", media.FileName)
	assert.Equal(t, "new-original.jpg", media.OriginalName)
	assert.Equal(t, "image/png", media.MimeType)
	assert.Equal(t, "/new/path", media.Path)
	assert.Equal(t, "https://new-url.com", media.URL)
	assert.Equal(t, int64(2048), media.Size)
	assert.True(t, media.UpdatedAt.After(originalUpdatedAt))
}

func TestMedia_UpdateMedia_PartialUpdate(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "image/jpeg", "/old/path", "https://old-url.com", 1024, uuid.New())
	originalFileName := media.FileName
	originalOriginalName := media.OriginalName
	originalMimeType := media.MimeType
	originalPath := media.Path
	originalURL := media.URL

	err := media.UpdateMedia("", "", "", "", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName)         // Should remain unchanged
	assert.Equal(t, originalOriginalName, media.OriginalName) // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType)         // Should remain unchanged
	assert.Equal(t, originalPath, media.Path)                 // Should remain unchanged
	assert.Equal(t, originalURL, media.URL)                   // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlyFileName(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "image/jpeg", "/old/path", "https://old-url.com", 1024, uuid.New())
	originalOriginalName := media.OriginalName
	originalMimeType := media.MimeType
	originalPath := media.Path
	originalURL := media.URL
	originalSize := media.Size

	err := media.UpdateMedia("new-file.jpg", "", "", "", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, "new-file.jpg", media.FileName)
	assert.Equal(t, originalOriginalName, media.OriginalName) // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType)         // Should remain unchanged
	assert.Equal(t, originalPath, media.Path)                 // Should remain unchanged
	assert.Equal(t, originalURL, media.URL)                   // Should remain unchanged
	assert.Equal(t, originalSize, media.Size)                 // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlySize(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "image/jpeg", "/old/path", "https://old-url.com", 1024, uuid.New())
	originalFileName := media.FileName
	originalOriginalName := media.OriginalName
	originalMimeType := media.MimeType
	originalPath := media.Path
	originalURL := media.URL

	err := media.UpdateMedia("", "", "", "", "", 2048)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName)         // Should remain unchanged
	assert.Equal(t, originalOriginalName, media.OriginalName) // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType)         // Should remain unchanged
	assert.Equal(t, originalPath, media.Path)                 // Should remain unchanged
	assert.Equal(t, originalURL, media.URL)                   // Should remain unchanged
	assert.Equal(t, int64(2048), media.Size)
}

func TestMedia_SoftDelete(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "image/jpeg", "/test/path", "https://test-url.com", 1024, uuid.New())
	originalUpdatedAt := media.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	media.SoftDelete()

	assert.NotNil(t, media.DeletedAt)
	assert.True(t, media.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, media.IsDeleted())
}

func TestMedia_IsDeleted(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "image/jpeg", "/test/path", "https://test-url.com", 1024, uuid.New())

	// Initially not deleted
	assert.False(t, media.IsDeleted())

	// After soft delete
	media.SoftDelete()
	assert.True(t, media.IsDeleted())
}

func TestMedia_IsImage(t *testing.T) {
	media, _ := entities.NewMedia("test-image.jpg", "test-original.jpg", "image/jpeg", "/test/path", "https://test-url.com", 1024, uuid.New())

	// Should be an image
	assert.True(t, media.IsImage())

	// Change to non-image
	media.MimeType = "application/pdf"
	assert.False(t, media.IsImage())

	// Change to another image type
	media.MimeType = "image/png"
	assert.True(t, media.IsImage())
}

func TestMedia_IsVideo(t *testing.T) {
	media, _ := entities.NewMedia("test-video.mp4", "test-original.mp4", "video/mp4", "/test/path", "https://test-url.com", 1024, uuid.New())

	// Should be a video
	assert.True(t, media.IsVideo())

	// Change to non-video
	media.MimeType = "image/jpeg"
	assert.False(t, media.IsVideo())

	// Change to another video type
	media.MimeType = "video/avi"
	assert.True(t, media.IsVideo())
}

func TestMedia_IsAudio(t *testing.T) {
	media, _ := entities.NewMedia("test-audio.mp3", "test-original.mp3", "audio/mpeg", "/test/path", "https://test-url.com", 1024, uuid.New())

	// Should be an audio file
	assert.True(t, media.IsAudio())

	// Change to non-audio
	media.MimeType = "image/jpeg"
	assert.False(t, media.IsAudio())

	// Change to another audio type
	media.MimeType = "audio/wav"
	assert.True(t, media.IsAudio())
}

func TestMedia_GetFileExtension(t *testing.T) {
	media, _ := entities.NewMedia("test-image.jpg", "test-original.jpg", "image/jpeg", "/test/path", "https://test-url.com", 1024, uuid.New())

	// Should return .jpg
	assert.Equal(t, ".jpg", media.GetFileExtension())

	// Change original name
	media.OriginalName = "test-document.pdf"
	assert.Equal(t, ".pdf", media.GetFileExtension())

	// No extension
	media.OriginalName = "testfile"
	assert.Equal(t, "", media.GetFileExtension())

	// Multiple dots
	media.OriginalName = "test.file.name.txt"
	assert.Equal(t, ".txt", media.GetFileExtension())

	// Empty original name
	media.OriginalName = ""
	assert.Equal(t, "", media.GetFileExtension())
}

func TestMedia_SoftDelete_MultipleCalls(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "image/jpeg", "/test/path", "https://test-url.com", 1024, uuid.New())

	// First soft delete
	media.SoftDelete()
	firstDeletedAt := media.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	media.SoftDelete()
	secondDeletedAt := media.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, media.IsDeleted())
}

func TestMedia_UpdateMedia_OnlyMimeType(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "image/jpeg", "/test/path", "https://test-url.com", 1024, uuid.New())
	originalFileName := media.FileName
	originalOriginalName := media.OriginalName
	originalPath := media.Path
	originalURL := media.URL
	originalSize := media.Size

	err := media.UpdateMedia("", "", "image/png", "", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName)         // Should remain unchanged
	assert.Equal(t, originalOriginalName, media.OriginalName) // Should remain unchanged
	assert.Equal(t, "image/png", media.MimeType)
	assert.Equal(t, originalPath, media.Path) // Should remain unchanged
	assert.Equal(t, originalURL, media.URL)   // Should remain unchanged
	assert.Equal(t, originalSize, media.Size) // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlyPath(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "image/jpeg", "/old/path", "https://test-url.com", 1024, uuid.New())
	originalFileName := media.FileName
	originalOriginalName := media.OriginalName
	originalMimeType := media.MimeType
	originalURL := media.URL
	originalSize := media.Size

	err := media.UpdateMedia("", "", "", "/new/path", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName)         // Should remain unchanged
	assert.Equal(t, originalOriginalName, media.OriginalName) // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType)         // Should remain unchanged
	assert.Equal(t, "/new/path", media.Path)
	assert.Equal(t, originalURL, media.URL)   // Should remain unchanged
	assert.Equal(t, originalSize, media.Size) // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlyURL(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "image/jpeg", "/test/path", "https://old-url.com", 1024, uuid.New())
	originalFileName := media.FileName
	originalOriginalName := media.OriginalName
	originalMimeType := media.MimeType
	originalPath := media.Path
	originalSize := media.Size

	err := media.UpdateMedia("", "", "", "", "https://new-url.com", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName)         // Should remain unchanged
	assert.Equal(t, originalOriginalName, media.OriginalName) // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType)         // Should remain unchanged
	assert.Equal(t, originalPath, media.Path)                 // Should remain unchanged
	assert.Equal(t, "https://new-url.com", media.URL)
	assert.Equal(t, originalSize, media.Size) // Should remain unchanged
}
