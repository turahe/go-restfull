package model

import "time"

type UserRole struct {
	ID     uint `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID uint `json:"user_id" gorm:"not null;index"`
	RoleID uint `json:"role_id" gorm:"not null;index"`

	CreatedAt time.Time `json:"created_at"`
}

