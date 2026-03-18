package model

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"type:varchar(100);not null"`
	Slug string `json:"slug" gorm:"type:varchar(120);not null;uniqueIndex"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

