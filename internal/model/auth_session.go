package model

import (
	"time"

	"gorm.io/gorm"
)

// AuthSession represents a login session bound to a device.
type AuthSession struct {
	ID        string `json:"id" gorm:"primaryKey;type:char(36)"`
	UserID    uint   `json:"user_id" gorm:"not null;index"`
	DeviceID  string `json:"device_id" gorm:"type:varchar(64);not null;index"`
	IPAddress string `json:"ip_address" gorm:"type:varchar(45);not null"`
	UserAgent string `json:"user_agent" gorm:"type:varchar(255);not null"`

	RevokedAt *time.Time `json:"revoked_at,omitempty" gorm:"index"`
	RevokedBy *uint      `json:"revoked_by,omitempty" gorm:"index"`

	LastSeenAt time.Time      `json:"last_seen_at" gorm:"index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

