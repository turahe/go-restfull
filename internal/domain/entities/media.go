package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Media represents the core media domain entity
type Media struct {
	ID           uuid.UUID  `json:"id"`
	FileName     string     `json:"file_name"`
	OriginalName string     `json:"original_name"`
	MimeType     string     `json:"mime_type"`
	Size         int64      `json:"size"`
	Path         string     `json:"path"`
	URL          string     `json:"url"`
	UserID       uuid.UUID  `json:"user_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// NewMedia creates a new media with validation
func NewMedia(fileName, originalName, mimeType, path, url string, size int64, userID uuid.UUID) (*Media, error) {
	if fileName == "" {
		return nil, errors.New("file_name is required")
	}
	if originalName == "" {
		return nil, errors.New("original_name is required")
	}
	if mimeType == "" {
		return nil, errors.New("mime_type is required")
	}
	if path == "" {
		return nil, errors.New("path is required")
	}
	if url == "" {
		return nil, errors.New("url is required")
	}
	if size <= 0 {
		return nil, errors.New("size must be greater than 0")
	}
	if userID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}

	now := time.Now()
	return &Media{
		ID:           uuid.New(),
		FileName:     fileName,
		OriginalName: originalName,
		MimeType:     mimeType,
		Size:         size,
		Path:         path,
		URL:          url,
		UserID:       userID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateMedia updates media information
func (m *Media) UpdateMedia(fileName, originalName, mimeType, path, url string, size int64) error {
	if fileName != "" {
		m.FileName = fileName
	}
	if originalName != "" {
		m.OriginalName = originalName
	}
	if mimeType != "" {
		m.MimeType = mimeType
	}
	if path != "" {
		m.Path = path
	}
	if url != "" {
		m.URL = url
	}
	if size > 0 {
		m.Size = size
	}
	m.UpdatedAt = time.Now()
	return nil
}

// SoftDelete marks the media as deleted
func (m *Media) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
}

// IsDeleted checks if the media is deleted
func (m *Media) IsDeleted() bool {
	return m.DeletedAt != nil
}

// IsImage checks if the media is an image
func (m *Media) IsImage() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "image"
}

// IsVideo checks if the media is a video
func (m *Media) IsVideo() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "video"
}

// IsAudio checks if the media is an audio file
func (m *Media) IsAudio() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "audio"
}

// GetFileExtension returns the file extension
func (m *Media) GetFileExtension() string {
	if m.OriginalName == "" {
		return ""
	}

	for i := len(m.OriginalName) - 1; i >= 0; i-- {
		if m.OriginalName[i] == '.' {
			return m.OriginalName[i:]
		}
	}
	return ""
}
