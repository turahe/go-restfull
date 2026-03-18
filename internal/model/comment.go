package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	PostID  uint   `json:"post_id" gorm:"not null;index"`
	UserID  uint   `json:"user_id" gorm:"not null;index"`
	Content string `json:"content" gorm:"type:text;not null"`
	User    *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	Tags []Tag `json:"tags,omitempty" gorm:"many2many:comment_tags"`

	CreatedBy uint  `json:"created_by" gorm:"not null;index"`
	UpdatedBy uint  `json:"updated_by" gorm:"not null;index"`
	DeletedBy *uint `json:"deleted_by,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

