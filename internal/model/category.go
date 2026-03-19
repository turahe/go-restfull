package model

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"type:varchar(100);not null"`
	Slug string `json:"slug" gorm:"type:varchar(120);not null;uniqueIndex"`

	Posts []Post `json:"posts,omitempty"`

	Media []Media `json:"media,omitempty" gorm:"many2many:category_media;"`

	CreatedBy uint  `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint  `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint `json:"deletedBy,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (Category) TableName() string {
	return "categories"
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Category) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}
