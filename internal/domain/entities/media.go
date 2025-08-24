// Package entities provides the core domain models and business logic entities
// for the application. This file contains the Media entity for managing
// media files with support for hierarchical organization and file type detection.
package entities

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Media group constants for common use cases
const (
	MediaGroupAvatar   = "avatar"
	MediaGroupCover    = "cover"
	MediaGroupGallery  = "gallery"
	MediaGroupDocument = "document"
	MediaGroupVideo    = "video"
	MediaGroupAudio    = "audio"
)

// Media represents the core media domain entity that manages file metadata
// and organization. It supports hierarchical organization through nested set
// model fields and provides file type detection capabilities.
//
// The entity includes:
// - File metadata (name, filename, hash, size, MIME type)
// - Storage information (disk location)
// - Hierarchical organization support (nested set model)
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for data retention
type Media struct {
	ID             uuid.UUID  `json:"id"`                        // Unique identifier for the media
	Name           string     `json:"name"`                      // Display name for the media
	FileName       string     `json:"file_name"`                 // Original filename of the uploaded file
	Hash           string     `json:"hash"`                      // File hash for integrity verification
	Disk           string     `json:"disk"`                      // Storage disk identifier (local, s3, etc.)
	MimeType       string     `json:"mime_type"`                 // MIME type of the file for content identification
	Size           int64      `json:"size"`                      // File size in bytes
	RecordLeft     *uint64    `json:"record_left,omitempty"`     // Left boundary for nested set model
	RecordRight    *uint64    `json:"record_right,omitempty"`    // Right boundary for nested set model
	RecordOrdering *uint64    `json:"record_ordering,omitempty"` // Display order within the same level
	RecordDepth    *uint64    `json:"record_depth,omitempty"`    // Depth level in the hierarchy
	CreatedBy      uuid.UUID  `json:"created_by"`                // ID of user who uploaded this media
	UpdatedBy      uuid.UUID  `json:"updated_by"`                // ID of user who last updated this media
	DeletedBy      *uuid.UUID `json:"deleted_by,omitempty"`      // ID of user who deleted this media (soft delete)
	CreatedAt      time.Time  `json:"created_at"`                // Timestamp when media was uploaded
	UpdatedAt      time.Time  `json:"updated_at"`                // Timestamp when media was last updated
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`      // Timestamp when media was soft deleted
}

// NewMedia creates a new media entity with validation.
// This constructor validates required fields and initializes the media
// with generated UUID and timestamps.
//
// Parameters:
//   - name: Display name for the media
//   - fileName: Original filename of the uploaded file
//   - hash: File hash for integrity verification
//   - disk: Storage disk identifier
//   - mimeType: MIME type of the file
//   - size: File size in bytes
//   - userID: UUID of the user uploading the media
//
// Returns:
//   - *Media: Pointer to the newly created media entity
//   - error: Validation error if any required field is invalid
//
// Validation rules:
// - fileName, hash, mimeType, and disk cannot be empty
// - size must be greater than 0
// - userID must be a valid UUID
func NewMedia(name, fileName, hash, disk, mimeType string, size int64, userID uuid.UUID) (*Media, error) {
	// Validate required fields
	if fileName == "" {
		return nil, errors.New("file_name is required")
	}
	if hash == "" {
		return nil, errors.New("hash is required")
	}
	if mimeType == "" {
		return nil, errors.New("mime_type is required")
	}
	if disk == "" {
		return nil, errors.New("disk is required")
	}
	if size <= 0 {
		return nil, errors.New("size must be greater than 0")
	}
	if userID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}

	// Create media with current timestamp
	now := time.Now()
	return &Media{
		ID:        uuid.New(), // Generate new unique identifier
		Name:      name,       // Set display name
		FileName:  fileName,   // Set original filename
		Hash:      hash,       // Set file hash
		Disk:      disk,       // Set storage disk
		MimeType:  mimeType,   // Set MIME type
		Size:      size,       // Set file size
		CreatedAt: now,        // Set creation timestamp
		UpdatedAt: now,        // Set initial update timestamp
	}, nil
}

// UpdateMedia updates media information with new values.
// This method allows partial updates - only non-empty values are updated.
// The UpdatedAt timestamp is automatically updated when any field changes.
//
// Parameters:
//   - name: New display name (optional, only updated if not empty)
//   - fileName: New filename (optional, only updated if not empty)
//   - hash: New file hash (optional, only updated if not empty)
//   - disk: New storage disk (optional, only updated if not empty)
//   - mimeType: New MIME type (optional, only updated if not empty)
//   - size: New file size (optional, only updated if greater than 0)
//
// Returns:
//   - error: Always nil, included for interface consistency
//
// Note: This method automatically updates the UpdatedAt timestamp
func (m *Media) UpdateMedia(name, fileName, hash, disk, mimeType string, size int64) error {
	// Update fields only if new values are provided
	if fileName != "" {
		m.FileName = fileName
	}
	if hash != "" {
		m.Hash = hash
	}
	if mimeType != "" {
		m.MimeType = mimeType
	}
	if disk != "" {
		m.Disk = disk
	}
	if size > 0 {
		m.Size = size
	}

	// Update modification timestamp
	m.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the media as deleted without removing it from the database.
// This sets the DeletedAt timestamp and updates the UpdatedAt timestamp.
// The media will be excluded from normal queries but remains accessible
// for audit and recovery purposes.
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (m *Media) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now // Set deletion timestamp
	m.UpdatedAt = now  // Update modification timestamp
}

// IsDeleted checks if the media has been soft deleted.
// Returns true if DeletedAt is not nil, false otherwise.
// This method is useful for filtering out deleted media from queries.
//
// Returns:
//   - bool: true if media is deleted, false if active
func (m *Media) IsDeleted() bool {
	return m.DeletedAt != nil
}

// IsImage checks if the media is an image file.
// This method examines the MIME type to determine if the file
// is an image format (starts with "image/").
//
// Returns:
//   - bool: true if the file is an image, false otherwise
//
// Note: Checks if MIME type starts with "image/" prefix
func (m *Media) IsImage() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "image"
}

// IsVideo checks if the media is a video file.
// This method examines the MIME type to determine if the file
// is a video format (starts with "video/").
//
// Returns:
//   - bool: true if the file is a video, false otherwise
//
// Note: Checks if MIME type starts with "video/" prefix
func (m *Media) IsVideo() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "video"
}

// IsAudio checks if the media is an audio file.
// This method examines the MIME type to determine if the file
// is an audio format (starts with "audio/").
//
// Returns:
//   - bool: true if the file is an audio file, false otherwise
//
// Note: Checks if MIME type starts with "audio/" prefix
func (m *Media) IsAudio() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "audio"
}

// GetFileExtension returns the file extension from the filename.
// This method extracts the extension by finding the last dot in the filename
// and returning everything from that dot onwards.
//
// Returns:
//   - string: File extension including the dot (e.g., ".jpg", ".pdf"), or empty string if no extension
//
// Examples:
//   - "image.jpg" returns ".jpg"
//   - "document.pdf" returns ".pdf"
//   - "noextension" returns ""
func (m *Media) GetFileExtension() string {
	if m.FileName == "" {
		return ""
	}

	// Find the last dot in the filename
	for i := len(m.FileName) - 1; i >= 0; i-- {
		if m.FileName[i] == '.' {
			return m.FileName[i:] // Return everything from the dot onwards
		}
	}
	return "" // No extension found
}

func (m *Media) GetURL() string {
	return fmt.Sprintf("%s/%s", m.Disk, m.FileName)
}

// IsAvatar checks if the media is marked as an avatar
func (m *Media) IsAvatar() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "image"
}

// GetFileSizeInMB returns the file size in megabytes
func (m *Media) GetFileSizeInMB() float64 {
	return float64(m.Size) / (1024 * 1024)
}

// GetFileSizeInKB returns the file size in kilobytes
func (m *Media) GetFileSizeInKB() float64 {
	return float64(m.Size) / 1024
}
