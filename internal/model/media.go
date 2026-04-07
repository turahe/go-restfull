package model

import (
	"time"

	"gorm.io/gorm"
)

// Media is user-scoped storage; folders use media_type "folder", files use image/file.
// Nested set is scoped by user_id (same pattern as categories, per-user forest).
type Media struct {
	ID     uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID uint   `json:"userId" gorm:"not null;uniqueIndex:idx_media_user_parent_name"`
	Name   string `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:idx_media_user_parent_name"`
	// ParentID nil for roots in this user's tree.
	ParentID *uint `json:"parentId,omitempty" gorm:"column:parent_id;uniqueIndex:idx_media_user_parent_name"`

	Lft   int `json:"lft" gorm:"column:lft;index;not null"`
	Rgt   int `json:"rgt" gorm:"column:rgt;index;not null"`
	Depth int `json:"depth" gorm:"column:depth;not null"`

	MediaType    string `json:"mediaType" gorm:"type:varchar(20);not null"` // "image", "file", or "folder"
	OriginalName string `json:"originalName" gorm:"type:varchar(255);not null"`
	MimeType     string `json:"mimeType" gorm:"type:varchar(100);not null"`
	Size         int64  `json:"size" gorm:"not null"`
	StoragePath  string `json:"storagePath" gorm:"type:varchar(512);not null;index;default:''"`

	DownloadURL string `json:"downloadUrl,omitempty" gorm:"-"`

	CreatedBy uint           `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint           `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint          `json:"deletedBy,omitempty" gorm:"index"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`

	CreatedAt time.Time `json:"createdAt" gorm:"index"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Media) TableName() string {
	return "media"
}

func (m *Media) BeforeCreate(tx *gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Media) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
