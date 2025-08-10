package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Media represents the core media domain entity
type Media struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	FileName       string     `json:"file_name"`
	Hash           string     `json:"hash"`
	Disk           string     `json:"disk"`
	MimeType       string     `json:"mime_type"`
	Size           int64      `json:"size"`
	RecordLeft     *uint64    `json:"record_left,omitempty"`
	RecordRight    *uint64    `json:"record_right,omitempty"`
	RecordOrdering *uint64    `json:"record_ordering,omitempty"`
	RecordDepth    *uint64    `json:"record_depth,omitempty"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	UpdatedBy      uuid.UUID  `json:"updated_by"`
	DeletedBy      *uuid.UUID `json:"deleted_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewMedia creates a new media with validation
func NewMedia(name, fileName, hash, disk, mimeType string, size int64, userID uuid.UUID) (*Media, error) {
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

	now := time.Now()
	return &Media{
		ID:        uuid.New(),
		Name:      name,
		FileName:  fileName,
		Hash:      hash,
		Disk:      disk,
		MimeType:  mimeType,
		Size:      size,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateMedia updates media information
func (m *Media) UpdateMedia(name, fileName, hash, disk, mimeType string, size int64) error {
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
	if m.FileName == "" {
		return ""
	}

	for i := len(m.FileName) - 1; i >= 0; i-- {
		if m.FileName[i] == '.' {
			return m.FileName[i:]
		}
	}
	return ""
}
