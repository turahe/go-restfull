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

	CreatedBy uint  `json:"createdBy" gorm:"not null;index"`
	UpdatedBy uint  `json:"updatedBy" gorm:"not null;index"`
	DeletedBy *uint `json:"deletedBy,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

