package model

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Key  string `json:"key" gorm:"type:varchar(200);not null;uniqueIndex"` // e.g. "/api/v1/posts:POST"
	Desc string `json:"desc,omitempty" gorm:"type:varchar(255)"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

