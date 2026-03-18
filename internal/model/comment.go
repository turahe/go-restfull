package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	PostID  uint   `json:"postId" gorm:"not null;index"`
	UserID  uint   `json:"userId" gorm:"not null;index"`
	Content string `json:"content" gorm:"type:text;not null"`
	User    *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	Tags []Tag `json:"tags,omitempty" gorm:"many2many:comment_tags"`

	Media []Media `json:"media,omitempty" gorm:"many2many:comment_media;"`

	CreatedBy uint  `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint  `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint `json:"deletedBy,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}
