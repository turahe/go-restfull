package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null"`
	// Unique index on slug also serves as the required index for lookups.
	Slug    string `json:"slug" gorm:"type:varchar(220);not null;uniqueIndex"`
	Content string `json:"content" gorm:"type:longtext;not null"`

	UserID uint  `json:"user_id" gorm:"not null;index"`
	User   *User `json:"author,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	CreatedBy uint  `json:"created_by" gorm:"not null;index"`
	UpdatedBy uint  `json:"updated_by" gorm:"not null;index"`
	DeletedBy *uint `json:"deleted_by,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Categories []Category `json:"categories,omitempty" gorm:"many2many:post_categories"`
}

