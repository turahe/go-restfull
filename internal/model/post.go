package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID    uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Title string `json:"title" gorm:"type:varchar(200);not null"`
	// Unique index on slug also serves as the required index for lookups.
	Slug    string `json:"slug" gorm:"type:varchar(220);not null;uniqueIndex"`
	Content string `json:"content" gorm:"type:longtext;not null"`

	UserID uint  `json:"userId" gorm:"not null;index"`
	User   *User `json:"author,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	CategoryID uint      `json:"categoryId" gorm:"not null;index"`
	Category   *Category `json:"category,omitempty" gorm:"constraint:OnDelete:RESTRICT"`

	Media []Media `json:"media,omitempty" gorm:"many2many:post_media;"`

	Tags []Tag `json:"tags,omitempty" gorm:"many2many:post_tags"`

	CreatedBy uint  `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint  `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint `json:"deletedBy,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"createdAt" gorm:"index"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (Post) TableName() string {
	return "posts"
}

func (p *Post) BeforeCreate(tx *gorm.DB) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Post) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}
