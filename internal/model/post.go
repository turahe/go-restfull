package model

import "time"

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null"`
	Slug      string    `json:"slug" gorm:"type:varchar(220);not null;uniqueIndex;index"`
	Content   string    `json:"content" gorm:"type:longtext;not null"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	User      *User     `json:"author,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
}

