package database

import (
	"go-rest/internal/model"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Post{},
		&model.Comment{},
	)
}

