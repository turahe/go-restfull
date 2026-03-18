package model

import "time"

type RolePermission struct {
	ID           uint `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID       uint `json:"role_id" gorm:"not null;index"`
	PermissionID uint `json:"permission_id" gorm:"not null;index"`

	CreatedAt time.Time `json:"created_at"`
}

