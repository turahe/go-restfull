package model

import (
	"time"

	"gorm.io/gorm"
)

type RolePermission struct {
	ID           uint `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID       uint `json:"roleId" gorm:"not null;index"`
	PermissionID uint `json:"permissionId" gorm:"not null;index"`

	CreatedAt time.Time `json:"createdAt"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (rp *RolePermission) BeforeCreate(tx *gorm.DB) error {
	rp.CreatedAt = time.Now()
	return nil
}
