package model

import "time"

type RolePermission struct {
	ID           uint `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID       uint `json:"roleId" gorm:"not null;index"`
	PermissionID uint `json:"permissionId" gorm:"not null;index"`

	CreatedAt time.Time `json:"createdAt"`
}

