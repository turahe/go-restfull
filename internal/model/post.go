package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PostLayout string

const (
	PostLayoutSimple PostLayout = "simple"
	PostLayoutAuthor PostLayout = "author"
	PostLayoutBook   PostLayout = "book"
	PostLayoutList   PostLayout = "list"
)

func (l PostLayout) IsValid() bool {
	switch l {
	case PostLayoutSimple, PostLayoutAuthor, PostLayoutBook, PostLayoutList:
		return true
	default:
		return false
	}
}

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

func (s PostStatus) IsValid() bool {
	switch s {
	case PostStatusDraft, PostStatusPublished, PostStatusArchived:
		return true
	default:
		return false
	}
}

type Post struct {
	ID    uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Title string `json:"title" gorm:"type:varchar(200);not null"`
	// Unique index on slug also serves as the required index for lookups.
	Slug    string `json:"slug" gorm:"type:varchar(220);not null;uniqueIndex"`
	Content string `json:"content" gorm:"type:longtext;not null"`

	// SEO / sharing lives in post_seo (optional; see PostSEO).
	PostSEO *PostSEO `json:"seo,omitempty" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`

	UserID uint  `json:"userId" gorm:"not null;index"`
	User   *User `json:"author,omitempty" gorm:"constraint:OnDelete:CASCADE"`

	CategoryID uint       `json:"categoryId" gorm:"not null;index"`
	Layout     PostLayout `json:"layout" gorm:"type:varchar(50);not null;check:layout IN ('simple','author','book','list')"`
	Status     PostStatus `json:"status" gorm:"type:varchar(20);not null;default:published;index;check:status IN ('draft','published','archived')"`
	Category   *Category  `json:"category,omitempty" gorm:"constraint:OnDelete:RESTRICT"`

	Media []Media `json:"media,omitempty" gorm:"many2many:post_media;"`

	Tags []Tag `json:"tags,omitempty" gorm:"many2many:post_tags"`

	CreatedBy uint  `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint  `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint `json:"deletedBy,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"createdAt" gorm:"index"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.Layout == "" {
		p.Layout = PostLayoutSimple
	}
	if !p.Layout.IsValid() {
		return fmt.Errorf("invalid post layout: %q", p.Layout)
	}
	if p.Status == "" {
		p.Status = PostStatusPublished
	}
	if !p.Status.IsValid() {
		return fmt.Errorf("invalid post status: %q", p.Status)
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Post) BeforeUpdate(tx *gorm.DB) error {
	if p.Layout != "" && !p.Layout.IsValid() {
		return fmt.Errorf("invalid post layout: %q", p.Layout)
	}
	if p.Status != "" && !p.Status.IsValid() {
		return fmt.Errorf("invalid post status: %q", p.Status)
	}
	p.UpdatedAt = time.Now()
	return nil
}
