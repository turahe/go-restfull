package model

import (
	"time"

	"gorm.io/gorm"
)

// UserTwoFactor stores TOTP configuration for a user.
type UserTwoFactor struct {
	UserID     uint       `json:"userId" gorm:"primaryKey;index"`
	SecretEnc  string     `json:"-" gorm:"type:varbinary(255);not null"`
	Enabled    bool       `json:"enabled"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func (UserTwoFactor) TableName() string {
	return "user_two_factors"
}

func (ut *UserTwoFactor) BeforeCreate(tx *gorm.DB) error {
	ut.CreatedAt = time.Now()
	ut.UpdatedAt = time.Now()
	return nil
}

func (ut *UserTwoFactor) BeforeUpdate(tx *gorm.DB) error {
	ut.UpdatedAt = time.Now()
	return nil
}

// TwoFactorChallenge represents a pending login challenge for 2FA-enabled users.
type TwoFactorChallenge struct {
	ID         string     `json:"id" gorm:"primaryKey;type:char(36)"`
	UserID     uint       `json:"userId" gorm:"not null;index"`
	DeviceID   string     `json:"deviceId" gorm:"type:varchar(64);not null"`
	ExpiresAt  time.Time  `json:"expiresAt" gorm:"index"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty" gorm:"index"`
	Attempts   int        `json:"attempts" gorm:"not null;default:0"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (TwoFactorChallenge) TableName() string {
	return "two_factor_challenges"
}

func (tc *TwoFactorChallenge) BeforeCreate(tx *gorm.DB) error {
	tc.CreatedAt = time.Now()
	return nil
}
