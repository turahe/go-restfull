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

	CreatedBy uint  `json:"created_by" gorm:"not null;index"`
	UpdatedBy uint  `json:"updated_by" gorm:"not null;index"`
	DeletedBy *uint `json:"deleted_by,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

