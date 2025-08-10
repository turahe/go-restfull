package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewMedia_Success(t *testing.T) {
	fileName := "test-image.jpg"
	name := "original-image.jpg"
	hash := "1234567890"
	disk := "local"
	mimeType := "image/jpeg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, name, hash, disk, mimeType, size, userID)

	assert.NoError(t, err)
	assert.NotNil(t, media)
	assert.Equal(t, fileName, media.FileName)
	assert.Equal(t, name, media.Name)
	assert.Equal(t, hash, media.Hash)
	assert.Equal(t, disk, media.Disk)
	assert.Equal(t, mimeType, media.MimeType)
	assert.Equal(t, size, media.Size)
	assert.NotEqual(t, uuid.Nil, media.ID)
	assert.False(t, media.CreatedAt.IsZero())
	assert.False(t, media.UpdatedAt.IsZero())
	assert.Nil(t, media.DeletedAt)
}

func TestNewMedia_EmptyFileName(t *testing.T) {
	name := "original-image.jpg"
	hash := "1234567890"
	disk := "local"
	mimeType := "image/jpeg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia("", name, hash, disk, mimeType, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "file_name is required", err.Error())
}

func TestNewMedia_EmptyName(t *testing.T) {
	fileName := "test-image.jpg"
	mimeType := "image/jpeg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, "", "", "", mimeType, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "name is required", err.Error())
}

func TestNewMedia_EmptyMimeType(t *testing.T) {
	fileName := "test-image.jpg"
	name := "original-image.jpg"
	hash := "1234567890"
	disk := "local"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, name, hash, disk, "", size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "hash is required", err.Error())
}

func TestNewMedia_EmptyDisk(t *testing.T) {
	fileName := "test-image.jpg"
	name := "original-image.jpg"
	hash := "1234567890"
	disk := "local"
	mimeType := "image/jpeg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, name, hash, disk, mimeType, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "disk is required", err.Error())
}

func TestNewMedia_ZeroSize(t *testing.T) {
	fileName := "test-image.jpg"
	name := "original-image.jpg"
	hash := "1234567890"
	disk := "local"
	mimeType := "image/jpeg"
	size := int64(1024)
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, name, hash, disk, mimeType, size, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "size must be greater than 0", err.Error())
}

func TestNewMedia_NegativeSize(t *testing.T) {
	fileName := "test-image.jpg"
	name := "original-image.jpg"
	hash := "1234567890"
	disk := "local"
	mimeType := "image/jpeg"
	userID := uuid.New()

	media, err := entities.NewMedia(fileName, name, hash, disk, mimeType, 0, userID)

	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, "size must be greater than 0", err.Error())
}

func TestMedia_UpdateMedia(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalUpdatedAt := media.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := media.UpdateMedia("new-file.jpg", "new-original.jpg", "1234567890", "local", "image/png", 2048)

	assert.NoError(t, err)
	assert.Equal(t, "new-file.jpg", media.FileName)
	assert.Equal(t, "new-original.jpg", media.Name)
	assert.Equal(t, "image/png", media.MimeType)
	assert.Equal(t, int64(2048), media.Size)
	assert.True(t, media.UpdatedAt.After(originalUpdatedAt))
}

func TestMedia_UpdateMedia_PartialUpdate(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalFileName := media.FileName
	originalName := media.Name
	originalMimeType := media.MimeType
	originalHash := media.Hash
	originalDisk := media.Disk

	err := media.UpdateMedia("", "", "", "", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName) // Should remain unchanged
	assert.Equal(t, originalName, media.Name)         // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType) // Should remain unchanged
	assert.Equal(t, originalHash, media.Hash)         // Should remain unchanged
	assert.Equal(t, originalDisk, media.Disk)         // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlyFileName(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalName := media.Name
	originalMimeType := media.MimeType
	originalHash := media.Hash
	originalDisk := media.Disk
	originalSize := media.Size

	err := media.UpdateMedia("new-file.jpg", "", "", "", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, "new-file.jpg", media.FileName)
	assert.Equal(t, originalName, media.Name)         // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType) // Should remain unchanged
	assert.Equal(t, originalHash, media.Hash)         // Should remain unchanged
	assert.Equal(t, originalDisk, media.Disk)         // Should remain unchanged
	assert.Equal(t, originalSize, media.Size)         // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlySize(t *testing.T) {
	media, _ := entities.NewMedia("old-file.jpg", "old-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalFileName := media.FileName
	originalName := media.Name
	originalMimeType := media.MimeType
	originalHash := media.Hash
	originalDisk := media.Disk

	err := media.UpdateMedia("", "", "", "", "", 2048)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName) // Should remain unchanged
	assert.Equal(t, originalName, media.Name)         // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType) // Should remain unchanged
	assert.Equal(t, originalHash, media.Hash)         // Should remain unchanged
	assert.Equal(t, originalDisk, media.Disk)         // Should remain unchanged
	assert.Equal(t, int64(2048), media.Size)
}

func TestMedia_SoftDelete(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalUpdatedAt := media.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	media.SoftDelete()

	assert.NotNil(t, media.DeletedAt)
	assert.True(t, media.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, media.IsDeleted())
}

func TestMedia_IsDeleted(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())

	// Initially not deleted
	assert.False(t, media.IsDeleted())

	// After soft delete
	media.SoftDelete()
	assert.True(t, media.IsDeleted())
}

func TestMedia_IsImage(t *testing.T) {
	media, _ := entities.NewMedia("test-image.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())

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
	media, _ := entities.NewMedia("test-video.mp4", "test-original.mp4", "1234567890", "local", "video/mp4", 1024, uuid.New())

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
	media, _ := entities.NewMedia("test-audio.mp3", "test-original.mp3", "1234567890", "local", "audio/mpeg", 1024, uuid.New())

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
	media, _ := entities.NewMedia("test-image.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())

	// Should return .jpg
	assert.Equal(t, ".jpg", media.GetFileExtension())

	// Change original name
	media.Name = "test-document.pdf"
	assert.Equal(t, ".pdf", media.GetFileExtension())

	// No extension
	media.Name = "testfile"
	assert.Equal(t, "", media.GetFileExtension())

	// Multiple dots
	media.Name = "test.file.name.txt"
	assert.Equal(t, ".txt", media.GetFileExtension())

	// Empty original name
	media.Name = ""
	assert.Equal(t, "", media.GetFileExtension())
}

func TestMedia_SoftDelete_MultipleCalls(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())

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
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalFileName := media.FileName
	originalName := media.Name
	originalHash := media.Hash
	originalDisk := media.Disk
	originalSize := media.Size

	err := media.UpdateMedia("", "", "image/png", "", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName) // Should remain unchanged
	assert.Equal(t, originalName, media.Name)         // Should remain unchanged
	assert.Equal(t, "image/png", media.MimeType)
	assert.Equal(t, originalHash, media.Hash) // Should remain unchanged
	assert.Equal(t, originalDisk, media.Disk) // Should remain unchanged
	assert.Equal(t, originalSize, media.Size) // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlyPath(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalFileName := media.FileName
	originalName := media.Name
	originalMimeType := media.MimeType
	originalHash := media.Hash
	originalDisk := media.Disk
	originalSize := media.Size

	err := media.UpdateMedia("", "", "", "/new/path", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName) // Should remain unchanged
	assert.Equal(t, originalName, media.Name)         // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType) // Should remain unchanged
	assert.Equal(t, originalHash, media.Hash)         // Should remain unchanged
	assert.Equal(t, originalDisk, media.Disk)         // Should remain unchanged
	assert.Equal(t, originalSize, media.Size)         // Should remain unchanged
}

func TestMedia_UpdateMedia_OnlyURL(t *testing.T) {
	media, _ := entities.NewMedia("test-file.jpg", "test-original.jpg", "1234567890", "local", "image/jpeg", 1024, uuid.New())
	originalFileName := media.FileName
	originalName := media.Name
	originalMimeType := media.MimeType
	originalHash := media.Hash
	originalDisk := media.Disk
	originalSize := media.Size

	err := media.UpdateMedia("", "", "", "", "", 0)

	assert.NoError(t, err)
	assert.Equal(t, originalFileName, media.FileName) // Should remain unchanged
	assert.Equal(t, originalName, media.Name)         // Should remain unchanged
	assert.Equal(t, originalMimeType, media.MimeType) // Should remain unchanged
	assert.Equal(t, originalHash, media.Hash)         // Should remain unchanged
	assert.Equal(t, originalDisk, media.Disk)         // Should remain unchanged
	assert.Equal(t, originalSize, media.Size)         // Should remain unchanged
}
