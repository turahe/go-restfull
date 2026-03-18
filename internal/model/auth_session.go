package model

import (
	"time"

	"gorm.io/gorm"
)

// AuthSession represents a login session bound to a device.
type AuthSession struct {
	ID        string `json:"id" gorm:"primaryKey;type:char(36)"`
	UserID    uint   `json:"userId" gorm:"not null;index"`
	DeviceID  string `json:"deviceId" gorm:"type:varchar(64);not null;index"`
	IPAddress string `json:"ipAddress" gorm:"type:varchar(45);not null"`
	UserAgent string `json:"userAgent" gorm:"type:varchar(255);not null"`

	RevokedAt *time.Time `json:"revokedAt,omitempty" gorm:"index"`
	RevokedBy *uint      `json:"revokedBy,omitempty" gorm:"index"`

	LastSeenAt time.Time      `json:"lastSeenAt" gorm:"index"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

