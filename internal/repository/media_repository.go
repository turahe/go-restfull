package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func mediaUserLockName(userID uint) string {
	return fmt.Sprintf("go_restfull_media_user_%d", userID)
}

func lockMediaUser(tx *gorm.DB, userID uint) error {
	if tx.Dialector.Name() != "mysql" {
		return nil
	}
	return tx.Exec("SELECT GET_LOCK(?, -1)", mediaUserLockName(userID)).Error
}

func unlockMediaUser(tx *gorm.DB, userID uint) {
	if tx.Dialector.Name() != "mysql" {
		return
	}
	_ = tx.Exec("SELECT RELEASE_LOCK(?)", mediaUserLockName(userID)).Error
}

type MediaRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewMediaRepository(db *gorm.DB, log *zap.Logger) *MediaRepository {
	return &MediaRepository{db: db, log: log}
}

// CreateFolderRoot creates a folder as a new root in the user's tree (media_type "folder").
func (r *MediaRepository) CreateFolderRoot(ctx context.Context, userID uint, name string, actorUserID uint) (*model.Media, error) {
	name = strings.TrimSpace(name)
	var out *model.Media
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMediaUser(tx, userID); err != nil {
			return err
		}
		defer unlockMediaUser(tx, userID)

		var dup int64
		if err := tx.Model(&model.Media{}).Where("user_id = ? AND parent_id IS NULL AND name = ?", userID, name).Count(&dup).Error; err != nil {
			return err
		}
		if dup > 0 {
			return gorm.ErrDuplicatedKey
		}

		var n int64
		if err := tx.Model(&model.Media{}).Where("user_id = ?", userID).Count(&n).Error; err != nil {
			return err
		}

		var maxRgt int
		q := tx.Model(&model.Media{}).Where("user_id = ?", userID)
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.Select("COALESCE(MAX(rgt), 0)").Scan(&maxRgt).Error; err != nil {
			return err
		}

		var lft, rgt int
		if n == 0 {
			lft, rgt = 1, 2
		} else {
			lft = maxRgt + 1
			rgt = maxRgt + 2
		}

		out = &model.Media{
			UserID:       userID,
			Name:         name,
			Lft:          lft,
			Rgt:          rgt,
			Depth:        0,
			MediaType:    "folder",
			OriginalName: name,
			MimeType:     "application/x-directory",
			Size:         0,
			StoragePath:  "",
			CreatedBy:    actorUserID,
			UpdatedBy:    actorUserID,
		}
		return tx.Create(out).Error
	})
	if err != nil {
		r.log.Error("create root media folder failed", zap.Error(err))
		return nil, err
	}
	return out, nil
}

// CreateFolderChild creates a folder as the last child of parentID (must belong to userID).
func (r *MediaRepository) CreateFolderChild(ctx context.Context, userID uint, parentID uint, name string, actorUserID uint) (*model.Media, error) {
	name = strings.TrimSpace(name)
	var out *model.Media
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMediaUser(tx, userID); err != nil {
			return err
		}
		defer unlockMediaUser(tx, userID)

		var dup int64
		if err := tx.Model(&model.Media{}).Where("user_id = ? AND parent_id = ? AND name = ?", userID, parentID, name).Count(&dup).Error; err != nil {
			return err
		}
		if dup > 0 {
			return gorm.ErrDuplicatedKey
		}

		var parent model.Media
		q := tx.Where("user_id = ? AND id = ?", userID, parentID)
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&parent).Error; err != nil {
			return err
		}

		parentRgt := parent.Rgt
		if err := tx.Exec("UPDATE media SET rgt = rgt + 2 WHERE user_id = ? AND deleted_at IS NULL AND rgt >= ?", userID, parentRgt).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE media SET lft = lft + 2 WHERE user_id = ? AND deleted_at IS NULL AND lft > ?", userID, parentRgt).Error; err != nil {
			return err
		}

		pid := parentID
		out = &model.Media{
			UserID:       userID,
			ParentID:     &pid,
			Name:         name,
			Lft:          parentRgt,
			Rgt:          parentRgt + 1,
			Depth:        parent.Depth + 1,
			MediaType:    "folder",
			OriginalName: name,
			MimeType:     "application/x-directory",
			Size:         0,
			StoragePath:  "",
			CreatedBy:    actorUserID,
			UpdatedBy:    actorUserID,
		}
		return tx.Create(out).Error
	})
	if err != nil {
		r.log.Error("create child media folder failed", zap.Error(err))
		return nil, err
	}
	return out, nil
}

// CreateFileRoot inserts a file as a new root node (caller sets Name, OriginalName, MimeType, Size, StoragePath, MediaType).
func (r *MediaRepository) CreateFileRoot(ctx context.Context, m *model.Media) error {
	if m == nil || m.UserID == 0 {
		return errors.New("invalid media")
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMediaUser(tx, m.UserID); err != nil {
			return err
		}
		defer unlockMediaUser(tx, m.UserID)

		var dup int64
		if err := tx.Model(&model.Media{}).Where("user_id = ? AND parent_id IS NULL AND name = ?", m.UserID, m.Name).Count(&dup).Error; err != nil {
			return err
		}
		if dup > 0 {
			return gorm.ErrDuplicatedKey
		}

		var n int64
		if err := tx.Model(&model.Media{}).Where("user_id = ?", m.UserID).Count(&n).Error; err != nil {
			return err
		}

		var maxRgt int
		q := tx.Model(&model.Media{}).Where("user_id = ?", m.UserID)
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.Select("COALESCE(MAX(rgt), 0)").Scan(&maxRgt).Error; err != nil {
			return err
		}

		if n == 0 {
			m.Lft, m.Rgt = 1, 2
		} else {
			m.Lft = maxRgt + 1
			m.Rgt = maxRgt + 2
		}
		m.Depth = 0
		m.ParentID = nil
		return tx.Create(m).Error
	})
	if err != nil {
		r.log.Error("create root media file failed", zap.Error(err))
		return err
	}
	return nil
}

// CreateFileChild inserts a file under parentID (caller sets Name, OriginalName, etc.).
func (r *MediaRepository) CreateFileChild(ctx context.Context, userID uint, parentID uint, m *model.Media) error {
	if m == nil || userID == 0 || parentID == 0 {
		return errors.New("invalid media")
	}
	m.UserID = userID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMediaUser(tx, userID); err != nil {
			return err
		}
		defer unlockMediaUser(tx, userID)

		var dup int64
		if err := tx.Model(&model.Media{}).Where("user_id = ? AND parent_id = ? AND name = ?", userID, parentID, m.Name).Count(&dup).Error; err != nil {
			return err
		}
		if dup > 0 {
			return gorm.ErrDuplicatedKey
		}

		var parent model.Media
		q := tx.Where("user_id = ? AND id = ?", userID, parentID)
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&parent).Error; err != nil {
			return err
		}

		parentRgt := parent.Rgt
		if err := tx.Exec("UPDATE media SET rgt = rgt + 2 WHERE user_id = ? AND deleted_at IS NULL AND rgt >= ?", userID, parentRgt).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE media SET lft = lft + 2 WHERE user_id = ? AND deleted_at IS NULL AND lft > ?", userID, parentRgt).Error; err != nil {
			return err
		}

		pid := parentID
		m.ParentID = &pid
		m.Lft = parentRgt
		m.Rgt = parentRgt + 1
		m.Depth = parent.Depth + 1
		return tx.Create(m).Error
	})
	if err != nil {
		r.log.Error("create child media file failed", zap.Error(err))
		return err
	}
	return nil
}

// GetTree returns all media for a user ordered by lft.
func (r *MediaRepository) GetTree(ctx context.Context, userID uint) ([]model.Media, error) {
	var rows []model.Media
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("lft ASC").Find(&rows).Error; err != nil {
		r.log.Error("get media tree failed", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// GetSubtree returns media in the subtree rooted at mediaID for this user.
func (r *MediaRepository) GetSubtree(ctx context.Context, userID uint, mediaID uint) ([]model.Media, error) {
	var anchor model.Media
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, mediaID).First(&anchor).Error; err != nil {
		r.log.Error("get media subtree anchor failed", zap.Error(err))
		return nil, err
	}
	var rows []model.Media
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND lft BETWEEN ? AND ?", userID, anchor.Lft, anchor.Rgt).
		Order("lft ASC").
		Find(&rows).Error
	if err != nil {
		r.log.Error("get media subtree failed", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// UpdateName updates display name (nested-set indices unchanged).
func (r *MediaRepository) UpdateName(ctx context.Context, userID uint, id uint, name string, actorUserID uint) (*model.Media, error) {
	name = strings.TrimSpace(name)
	var m model.Media
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&m).Error; err != nil {
		r.log.Error("find media for update failed", zap.Error(err))
		return nil, err
	}

	dupQ := r.db.WithContext(ctx).Model(&model.Media{}).Where("user_id = ? AND name = ? AND id <> ?", userID, name, id)
	if m.ParentID == nil {
		dupQ = dupQ.Where("parent_id IS NULL")
	} else {
		dupQ = dupQ.Where("parent_id = ?", *m.ParentID)
	}
	var dup int64
	if err := dupQ.Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, gorm.ErrDuplicatedKey
	}

	now := time.Now()
	updates := map[string]interface{}{
		"name":       name,
		"updated_by": actorUserID,
		"updated_at": now,
	}
	if m.MediaType == "folder" {
		updates["original_name"] = name
	}
	if err := r.db.WithContext(ctx).Model(&model.Media{}).Where("user_id = ? AND id = ?", userID, id).Updates(updates).Error; err != nil {
		r.log.Error("update media name failed", zap.Error(err))
		return nil, err
	}
	return r.FindByIDAndUserID(ctx, id, userID)
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
		n := "%" + req.Name + "%"
		countQ = countQ.Where("(name LIKE ? OR original_name LIKE ?)", n, n)
	}
	if s := strings.TrimSpace(req.Search); s != "" {
		n := "%" + s + "%"
		countQ = countQ.Where("(name LIKE ? OR original_name LIKE ?)", n, n)
	}

	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count media", zap.Error(err))
		return CursorPage{}, err
	}

	dataQ := r.db.WithContext(ctx).Model(&model.Media{}).Where("user_id = ?", userID)
	if req.Name != "" {
		n := "%" + req.Name + "%"
		dataQ = dataQ.Where("(name LIKE ? OR original_name LIKE ?)", n, n)
	}
	if s := strings.TrimSpace(req.Search); s != "" {
		n := "%" + s + "%"
		dataQ = dataQ.Where("(name LIKE ? OR original_name LIKE ?)", n, n)
	}

	var rows []model.Media
	if err := dataQ.
		Order("lft asc").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		r.log.Error("failed to list media", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		return CursorPage{Items: []model.Media{}, NextCursor: nil, PrevCursor: nil}, nil
	}

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

// ListByUserID keeps older callers/tests working.
func (r *MediaRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]model.Media, error) {
	page, err := r.List(ctx, userID, request.MediaListRequest{
		PageRequest: request.PageRequest{Page: 1, Limit: limit},
	})
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

func deleteMediaJoinRows(tx *gorm.DB, mediaIDs []uint) error {
	if len(mediaIDs) == 0 {
		return nil
	}
	if err := tx.Exec("DELETE FROM post_media WHERE media_id IN ?", mediaIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM user_media WHERE media_id IN ?", mediaIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM category_media WHERE media_id IN ?", mediaIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM comment_media WHERE media_id IN ?", mediaIDs).Error; err != nil {
		return err
	}
	return nil
}

// DeleteSubtree soft-deletes a media node and all descendants for this user and rebalances nested-set indices.
func (r *MediaRepository) DeleteSubtree(ctx context.Context, userID uint, mediaID uint, deletedBy uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMediaUser(tx, userID); err != nil {
			return err
		}
		defer unlockMediaUser(tx, userID)

		var anchor model.Media
		q := tx.Where("user_id = ? AND id = ?", userID, mediaID)
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&anchor).Error; err != nil {
			return err
		}

		var ids []uint
		if err := tx.Model(&model.Media{}).Where("user_id = ? AND lft BETWEEN ? AND ?", userID, anchor.Lft, anchor.Rgt).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := deleteMediaJoinRows(tx, ids); err != nil {
			return err
		}

		width := anchor.Rgt - anchor.Lft + 1
		now := time.Now()
		if err := tx.Model(&model.Media{}).Where("user_id = ? AND lft BETWEEN ? AND ?", userID, anchor.Lft, anchor.Rgt).Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
			"updated_by": deletedBy,
		}).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE media SET rgt = rgt - ? WHERE user_id = ? AND deleted_at IS NULL AND rgt > ?", width, userID, anchor.Rgt).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE media SET lft = lft - ? WHERE user_id = ? AND deleted_at IS NULL AND lft > ?", width, userID, anchor.Rgt).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		r.log.Error("delete media subtree failed", zap.Error(err))
		return err
	}
	return nil
}

// SoftDeleteByID removes a single media row (or use DeleteSubtree for nested-set consistency — both work for a leaf).
func (r *MediaRepository) SoftDeleteByID(ctx context.Context, id uint, userID uint, deletedBy uint) error {
	return r.DeleteSubtree(ctx, userID, id, deletedBy)
}

// AttachMedia associates an existing Media record with a specific target.
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
		err := r.db.WithContext(ctx).Exec(
			"INSERT INTO category_media (category_id, media_id) VALUES (?, ?)",
			targetID, mediaID,
		).Error
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
		Where("user_media.user_id = ? AND user_media.type = ? AND media.media_type <> ?", user.ID, "avatar", "folder").
		Limit(1).
		First(&avatar).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fallback := "https://ui-avatars.com/api/?name=" + user.Name
			return &fallback, nil
		}
		return nil, err
	}

	if strings.TrimSpace(avatar.DownloadURL) == "" {
		fallback := "https://ui-avatars.com/api/?name=" + user.Name
		return &fallback, nil
	}

	return &avatar.DownloadURL, nil
}
