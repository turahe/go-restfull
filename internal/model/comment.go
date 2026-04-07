package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment is a threaded comment on a post (nested set scoped by post_id).
type Comment struct {
	ID     uint `json:"id" gorm:"primaryKey;autoIncrement"`
	PostID uint `json:"postId" gorm:"not null;index"`
	// ParentID nil for root comments in this post's tree.
	ParentID *uint `json:"parentId,omitempty" gorm:"column:parent_id;index"`

	Lft   int `json:"lft" gorm:"column:lft;index;not null"`
	Rgt   int `json:"rgt" gorm:"column:rgt;index;not null"`
	Depth int `json:"depth" gorm:"column:depth;not null"`

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

func (Comment) TableName() string {
	return "comments"
}

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Comment) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}
