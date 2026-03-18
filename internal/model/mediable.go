package model

import "gorm.io/gorm"

// Mediable is the polymorphic join table between Media and any "mediable" model.
//
// Columns:
// - media_id: refers to Media.ID
// - mediable_type: "Post" | "User" | "Category" | "Comment"
// - mediable_id: refers to the corresponding model's ID
//
// Note: This repo currently uses separate join tables for GORM relations.
// This model is added to create the requested table via AutoMigrate.
type Mediable struct {
	MediaID      uint           `gorm:"primaryKey;column:media_id"`
	MediableType string         `gorm:"primaryKey;type:varchar(50);column:mediable_type"`
	MediableID   uint           `gorm:"primaryKey;column:mediable_id"`
	DeletedAt     gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func (Mediable) TableName() string { return "mediable" }

