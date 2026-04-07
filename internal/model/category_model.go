package model

import (
	"time"

	"gorm.io/gorm"
)

// CategoryModel maps to table categories (nested set: lft, rgt, depth).
// Unique (parent_id, name) prevents duplicate sibling names; roots (parent_id NULL) also checked in application code where MySQL allows duplicate (NULL, name).
type CategoryModel struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:idx_categories_parent_name"`
	ParentID *uint  `json:"parentId,omitempty" gorm:"column:parent_id;index;uniqueIndex:idx_categories_parent_name"`

	Lft   int `json:"lft" gorm:"column:lft;index;not null"`
	Rgt   int `json:"rgt" gorm:"column:rgt;index;not null"`
	Depth int `json:"depth" gorm:"column:depth;not null"`

	CreatedBy uint  `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint  `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint `json:"deletedBy,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (CategoryModel) TableName() string {
	return "categories"
}
