package database

import (
	"github.com/turahe/go-restfull/internal/model"

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
		&model.CategoryModel{},
		&model.Tag{},
		&model.Post{},
		&model.PostSEO{},
		&model.PostMedia{},
		&model.PostTag{},
		&model.Comment{},
		&model.Media{},
		&model.UserMedia{},
		&model.Setting{},
	)
}
