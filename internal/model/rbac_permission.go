package model

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Key  string `json:"key" gorm:"type:varchar(200);not null;uniqueIndex"` // e.g. "/api/v1/posts:POST"
	Desc string `json:"desc,omitempty" gorm:"type:varchar(255)"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (Permission) TableName() string {
	return "permissions"
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Permission) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}
