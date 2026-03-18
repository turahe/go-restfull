package model

import (
	"time"

	"gorm.io/gorm"
)

type Media struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	UserID uint `json:"userId" gorm:"not null;index"`

	MediaType     string `json:"mediaType" gorm:"type:varchar(20);not null"` // "image" or "file"
	OriginalName  string `json:"originalName" gorm:"type:varchar(255);not null"`
	MimeType      string `json:"mimeType" gorm:"type:varchar(100);not null"`
	Size          int64  `json:"size" gorm:"not null"`
	StoragePath   string `json:"storagePath" gorm:"type:varchar(512);not null;index"` // relative path under upload dir

	// DownloadURL is returned to clients for convenience when using MinIO.
	// It's not persisted in the database.
	DownloadURL string `json:"downloadUrl,omitempty" gorm:"-"`

	CreatedBy uint      `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint      `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint     `json:"deletedBy,omitempty" gorm:"index"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"createdAt" gorm:"index"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

