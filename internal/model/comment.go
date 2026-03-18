package model

import "time"

type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	PostID    uint      `json:"post_id" gorm:"not null;index"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	User      *User     `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

