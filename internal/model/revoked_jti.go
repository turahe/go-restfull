package model

import "time"

// RevokedJTI is a blacklist for revoked access tokens by jti until expiry.
type RevokedJTI struct {
	JTI       string    `json:"jti" gorm:"primaryKey;type:char(36)"`
	UserID    uint      `json:"userId" gorm:"not null;index"`
	SessionID string    `json:"sessionId" gorm:"type:char(36);not null;index"`
	Reason    string    `json:"reason" gorm:"type:varchar(120);not null"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"index"`
	RevokedAt time.Time `json:"revokedAt" gorm:"autoCreateTime"`
}

