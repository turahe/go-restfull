package repository

import (
	"context"
	"errors"
	"strings"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MediaRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewMediaRepository(db *gorm.DB, log *zap.Logger) *MediaRepository {
	return &MediaRepository{db: db, log: log}
}

func (r *MediaRepository) Create(ctx context.Context, m *model.Media) error {
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		r.log.Error("failed to create media", zap.Error(err))
		return err
	}
	return nil
}

func (r *MediaRepository) FindByIDAndUserID(ctx context.Context, id uint, userID uint) (*model.Media, error) {
	var m model.Media
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&m).Error
	if err != nil {
		r.log.Error("failed to find media by id and user id", zap.Error(err))
		return nil, err
	}
	return &m, nil
}

func (r *MediaRepository) List(ctx context.Context, userID uint, req request.MediaListRequest) (CursorPage, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	countQ := r.db.WithContext(ctx).Model(&model.Media{}).Where("user_id = ?", userID)
	if req.Name != "" {
		// Treat `name` as a fuzzy match on the original file name.
		countQ = countQ.Where("original_name LIKE ?", "%"+req.Name+"%")
	}

	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count media", zap.Error(err))
		return CursorPage{}, err
	}

	dataQ := r.db.WithContext(ctx).Model(&model.Media{}).Where("user_id = ?", userID)
	if req.Name != "" {
		dataQ = dataQ.Where("original_name LIKE ?", "%"+req.Name+"%")
	}

	var rows []model.Media
	if err := dataQ.
		Order("id desc").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		r.log.Error("failed to list media", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		return CursorPage{Items: []model.Media{}, NextCursor: nil, PrevCursor: nil}, nil
	}

	// Next/prev based on classic offset/limit existence.
	var nextCursor *uint
	if int64(offset)+int64(limit) < totalRows {
		tmp := rows[len(rows)-1].ID
		nextCursor = &tmp
	}

	var prevCursor *uint
	if offset > 0 {
		tmp := rows[0].ID
		prevCursor = &tmp
	}

	return CursorPage{Items: rows, NextCursor: nextCursor, PrevCursor: prevCursor}, nil
}

// ListByUserID keeps older callers/tests working. It returns only items (no pagination metadata).
func (r *MediaRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]model.Media, error) {
	page, err := r.List(ctx, userID, request.MediaListRequest{Limit: limit, Page: 1})
	if err != nil {
		return nil, err
	}
	items, ok := page.Items.([]model.Media)
	if !ok {
		r.log.Error("failed to convert media items", zap.Error(errors.New("failed to convert media items")))
		return nil, nil
	}
	return items, nil
}

func (r *MediaRepository) SoftDeleteByID(ctx context.Context, id uint, userID uint, deletedBy uint) error {
	now := deletedBy
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Media{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("deleted_by", &now).
			Error; err != nil {
			r.log.Error("failed to update media deleted by", zap.Error(err))
			return err
		}

		// Remove join-table references so the deleted media does not remain attached.
		// (Join tables are not guaranteed to have FK constraints in MySQL.)
		if err := tx.Exec("DELETE FROM post_media WHERE media_id = ?", id).Error; err != nil {
			r.log.Error("failed to delete post media", zap.Error(err))
			return err
		}
		if err := tx.Exec("DELETE FROM user_media WHERE media_id = ?", id).Error; err != nil {
			r.log.Error("failed to delete user media", zap.Error(err))
			return err
		}
		if err := tx.Exec("DELETE FROM category_media WHERE media_id = ?", id).Error; err != nil {
			r.log.Error("failed to delete category media", zap.Error(err))
			return err
		}
		if err := tx.Exec("DELETE FROM comment_media WHERE media_id = ?", id).Error; err != nil {
			r.log.Error("failed to delete comment media", zap.Error(err))
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
func (r *MediaRepository) AttachMedia(ctx context.Context, mediaID uint, targetType string, targetID uint) error {
	if mediaID == 0 || targetID == 0 {
		return errors.New("invalid ids")
	}

	switch targetType {
	case "Post":
		target := model.Post{ID: targetID}
		err := r.db.WithContext(ctx).
			Model(&target).
			Association("Media").
			Append(&model.Media{ID: mediaID})
		if err != nil {
			r.log.Error("failed to append media to post", zap.Error(err))
			return err
		}
		return nil
	case "User":
		// user_media has a NOT NULL `type` column (e.g. "avatar").
		// When using Association.Append(), GORM won't populate that extra join field,
		// so we insert the join row explicitly.
		j := model.UserMedia{
			UserID:  targetID,
			MediaID: mediaID,
			Type:    "media",
		}
		if err := r.db.WithContext(ctx).FirstOrCreate(&j).Error; err != nil {
			r.log.Error("failed to create user_media join row", zap.Error(err))
			return err
		}
		return nil
	case "Category":
		target := model.Category{ID: targetID}
		err := r.db.WithContext(ctx).
			Model(&target).
			Association("Media").
			Append(&model.Media{ID: mediaID})
		if err != nil {
			r.log.Error("failed to append media to category", zap.Error(err))
			return err
		}
		return nil
	case "Comment":
		target := model.Comment{ID: targetID}
		err := r.db.WithContext(ctx).
			Model(&target).
			Association("Media").
			Append(&model.Media{ID: mediaID})
		if err != nil {
			r.log.Error("failed to append media to comment", zap.Error(err))
			return err
		}
		return nil
	default:
		r.log.Error("invalid targetType", zap.String("targetType", targetType))
		return errors.New("invalid mediaableType")
	}
}

func (r *MediaRepository) UserAvatar(ctx context.Context, user *model.User) (*string, error) {
	if user == nil {
		return nil, errors.New("invalid user id")
	}

	var avatar model.Media
	err := r.db.WithContext(ctx).
		Model(&model.Media{}).
		Joins("INNER JOIN user_media ON media.id = user_media.media_id").
		Where("user_media.user_id = ? AND user_media.type = ?", user.ID, "avatar").
		Limit(1).
		First(&avatar).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback avatar when a user hasn't uploaded an avatar.
			fallback := "https://ui-avatars.com/api/?name=" + user.Name
			return &fallback, nil
		}
		return nil, err
	}

	// DownloadURL is not persisted (gorm:"-"), so it may be empty even when the join row exists.
	// In that case, return the same fallback URL.
	if strings.TrimSpace(avatar.DownloadURL) == "" {
		fallback := "https://ui-avatars.com/api/?name=" + user.Name
		return &fallback, nil
	}

	return &avatar.DownloadURL, nil
}
