package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name" gorm:"type:varchar(100);not null"`
	Email    string `json:"email" gorm:"type:varchar(190);not null;uniqueIndex"`
	Password string `json:"-" gorm:"type:varchar(255);not null"`

	Media []Media `json:"media,omitempty" gorm:"many2many:user_media;"`
	Roles []Role  `json:"roles,omitempty" gorm:"many2many:user_roles;"`
	// Avatar *Media  `json:"avatar,omitempty" gorm:"foreignKey:ID;references:AvatarID"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

type UserMedia struct {
	UserID    uint      `json:"user_id" gorm:"primaryKey"`
	MediaID   uint      `json:"media_id" gorm:"primaryKey"`
	Media     Media     `json:"media" gorm:"foreignKey:MediaID"`
	Type      string    `json:"type" gorm:"type:varchar(50);not null"`
	CreatedAt time.Time `json:"createdAt"`
}

func (UserMedia) TableName() string {
	return "user_media"
}

func (um *UserMedia) BeforeCreate(tx *gorm.DB) error {
	um.CreatedAt = time.Now()
	return nil
}
