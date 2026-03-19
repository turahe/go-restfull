package model

import (
	"time"

	"gorm.io/gorm"
)

type UserRole struct {
	ID     uint `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID uint `json:"userId" gorm:"not null;index"`
	RoleID uint `json:"roleId" gorm:"not null;index"`

	CreatedAt time.Time `json:"createdAt"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (ur *UserRole) BeforeCreate(tx *gorm.DB) error {
	ur.CreatedAt = time.Now()
	return nil
}
