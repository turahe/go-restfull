package model

import (
	"time"
)

// RefreshToken stores rotated refresh tokens (hashed).
type RefreshToken struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID      string    `json:"sessionId" gorm:"type:char(36);not null;index"`
	UserID         uint      `json:"userId" gorm:"not null;index"`
	TokenHash      string    `json:"-" gorm:"type:char(64);not null;uniqueIndex"`
	TokenFamily    string    `json:"tokenFamily" gorm:"type:char(36);not null;index"`
	RotatedFromID  *uint     `json:"rotatedFromId,omitempty" gorm:"index"`
	ExpiresAt      time.Time `json:"expiresAt" gorm:"index"`
	UsedAt         *time.Time `json:"usedAt,omitempty" gorm:"index"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty" gorm:"index"`
	RevokedReason  string     `json:"revokedReason,omitempty" gorm:"type:varchar(120)"`
	CreatedAt      time.Time  `json:"createdAt"`
}

