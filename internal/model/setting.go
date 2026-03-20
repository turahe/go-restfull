package model

import (
	"time"

	"gorm.io/gorm"
)

// Setting is a key/value row for application settings stored in the database.
// Only rows with IsPublic=true are exposed on the unauthenticated GET /settings API.
type Setting struct {
	ID    uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Key   string `json:"key" gorm:"column:setting_key;type:varchar(190);not null;uniqueIndex"`
	Value string `json:"value" gorm:"type:text;not null"`
	// IsPublic controls visibility on the public settings endpoint.
	IsPublic bool `json:"isPublic" gorm:"not null;default:true;index"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Setting) TableName() string {
	return "settings"
}

func (s *Setting) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	return nil
}

func (s *Setting) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}
