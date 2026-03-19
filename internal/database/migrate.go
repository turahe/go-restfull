package database

import (
	"go-rest/internal/model"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.AuthSession{},
		&model.RefreshToken{},
		&model.RevokedJTI{},
		&model.ImpersonationAudit{},
		&model.UserTwoFactor{},
		&model.TwoFactorChallenge{},
		&model.Category{},
		&model.Tag{},
		&model.Post{},
		&model.Comment{},
		&model.Media{},
		&model.UserMedia{},
	)
}
