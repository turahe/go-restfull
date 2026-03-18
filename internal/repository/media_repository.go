package repository

import (
	"context"
	"errors"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, m *model.Media) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *MediaRepository) FindByIDAndUserID(ctx context.Context, id uint, userID uint) (*model.Media, error) {
	var m model.Media
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MediaRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]model.Media, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	var rows []model.Media
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id desc").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *MediaRepository) SoftDeleteByID(ctx context.Context, id uint, userID uint, deletedBy uint) error {
	now := deletedBy
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Media{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("deleted_by", &now).
			Error; err != nil {
			return err
		}

		// Remove join-table references so the deleted media does not remain attached.
		// (Join tables are not guaranteed to have FK constraints in MySQL.)
		if err := tx.Exec("DELETE FROM post_media WHERE media_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM user_media WHERE media_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM category_media WHERE media_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM comment_media WHERE media_id = ?", id).Error; err != nil {
			return err
		}

		return tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Media{}).Error
	})
}

// AttachMedia associates an existing Media record with a specific target.
// This uses join tables configured via gorm tags on target models:
// - post_media
// - user_media
// - category_media
// - comment_media
func (r *MediaRepository) AttachMedia(ctx context.Context, mediaID uint, mediaableType string, mediaableID uint) error {
	if mediaID == 0 || mediaableID == 0 {
		return errors.New("invalid ids")
	}

	switch mediaableType {
	case "Post":
		target := model.Post{ID: mediaableID}
		return r.db.WithContext(ctx).
			Model(&target).
			Association("Media").
			Append(&model.Media{ID: mediaID})
	case "User":
		target := model.User{ID: mediaableID}
		return r.db.WithContext(ctx).
			Model(&target).
			Association("Media").
			Append(&model.Media{ID: mediaID})
	case "Category":
		target := model.Category{ID: mediaableID}
		return r.db.WithContext(ctx).
			Model(&target).
			Association("Media").
			Append(&model.Media{ID: mediaID})
	case "Comment":
		target := model.Comment{ID: mediaableID}
		return r.db.WithContext(ctx).
			Model(&target).
			Association("Media").
			Append(&model.Media{ID: mediaID})
	default:
		return errors.New("invalid mediaableType")
	}
}

