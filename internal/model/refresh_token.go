package model

import (
	"time"
)

// RefreshToken stores rotated refresh tokens (hashed).
type RefreshToken struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID      string    `json:"session_id" gorm:"type:char(36);not null;index"`
	UserID         uint      `json:"user_id" gorm:"not null;index"`
	TokenHash      string    `json:"-" gorm:"type:char(64);not null;uniqueIndex"`
	TokenFamily    string    `json:"token_family" gorm:"type:char(36);not null;index"`
	RotatedFromID  *uint     `json:"rotated_from_id,omitempty" gorm:"index"`
	ExpiresAt      time.Time `json:"expires_at" gorm:"index"`
	UsedAt         *time.Time `json:"used_at,omitempty" gorm:"index"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty" gorm:"index"`
	RevokedReason  string     `json:"revoked_reason,omitempty" gorm:"type:varchar(120)"`
	CreatedAt      time.Time  `json:"created_at"`
}

