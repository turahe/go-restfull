package model

import "time"

type UserRole struct {
	ID     uint `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID uint `json:"userId" gorm:"not null;index"`
	RoleID uint `json:"roleId" gorm:"not null;index"`

	CreatedAt time.Time `json:"createdAt"`
}

