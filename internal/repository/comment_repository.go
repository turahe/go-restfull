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

// ErrCommentSubtreeHasMedia is returned when delete would remove comments that still have rows in comment_media.
var ErrCommentSubtreeHasMedia = errors.New("comments in subtree still have attached media")

func commentPostLockName(postID uint) string {
	return fmt.Sprintf("go_restfull_comments_post_%d", postID)
}

func lockCommentPost(tx *gorm.DB, postID uint) error {
	if tx.Dialector.Name() != "mysql" {
		return nil
	}
	return tx.Exec("SELECT GET_LOCK(?, -1)", commentPostLockName(postID)).Error
}

func unlockCommentPost(tx *gorm.DB, postID uint) {
	if tx.Dialector.Name() != "mysql" {
		return
	}
	_ = tx.Exec("SELECT RELEASE_LOCK(?)", commentPostLockName(postID)).Error
}

type CommentRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewCommentRepository(db *gorm.DB, log *zap.Logger) *CommentRepository {
	return &CommentRepository{db: db, log: log}
}

// CreateRoot inserts a root comment for a post (depth 0, parent_id NULL).
func (r *CommentRepository) CreateRoot(ctx context.Context, postID uint, userID uint, content string, actorUserID uint) (*model.Comment, error) {
	var out *model.Comment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockCommentPost(tx, postID); err != nil {
			return err
		}
		defer unlockCommentPost(tx, postID)

		var n int64
		if err := tx.Model(&model.Comment{}).Where("post_id = ?", postID).Count(&n).Error; err != nil {
			return err
		}

		var maxRgt int
		q := tx.Model(&model.Comment{}).Where("post_id = ?", postID)
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

		out = &model.Comment{
			PostID:    postID,
			UserID:    userID,
			Content:   content,
			Lft:       lft,
			Rgt:       rgt,
			Depth:     0,
			CreatedBy: actorUserID,
			UpdatedBy: actorUserID,
		}
		return tx.Create(out).Error
	})
	if err != nil {
		r.log.Error("create root comment failed", zap.Error(err))
		return nil, err
	}
	return out, nil
}

// CreateChild inserts a comment as the last child of parentID (must belong to postID).
func (r *CommentRepository) CreateChild(ctx context.Context, postID uint, parentID uint, userID uint, content string, actorUserID uint) (*model.Comment, error) {
	var out *model.Comment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockCommentPost(tx, postID); err != nil {
			return err
		}
		defer unlockCommentPost(tx, postID)

		var parent model.Comment
		q := tx.Where("post_id = ? AND id = ?", postID, parentID)
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&parent).Error; err != nil {
			return err
		}

		parentRgt := parent.Rgt
		if err := tx.Exec("UPDATE comments SET rgt = rgt + 2 WHERE post_id = ? AND deleted_at IS NULL AND rgt >= ?", postID, parentRgt).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE comments SET lft = lft + 2 WHERE post_id = ? AND deleted_at IS NULL AND lft > ?", postID, parentRgt).Error; err != nil {
			return err
		}

		pid := parentID
		out = &model.Comment{
			PostID:    postID,
			ParentID:  &pid,
			UserID:    userID,
			Content:   content,
			Lft:       parentRgt,
			Rgt:       parentRgt + 1,
			Depth:     parent.Depth + 1,
			CreatedBy: actorUserID,
			UpdatedBy: actorUserID,
		}
		return tx.Create(out).Error
	})
	if err != nil {
		r.log.Error("create child comment failed", zap.Error(err))
		return nil, err
	}
	return out, nil
}

func (r *CommentRepository) ReplaceTags(ctx context.Context, commentID uint, tags []model.Tag) error {
	c := model.Comment{ID: commentID}
	err := r.db.WithContext(ctx).Model(&c).Association("Tags").Replace(tags)
	if err != nil {
		r.log.Error("failed to replace tags", zap.Error(err))
		return err
	}
	return nil
}

// GetTree returns all comments for a post ordered by lft.
func (r *CommentRepository) GetTree(ctx context.Context, postID uint) ([]model.Comment, error) {
	var rows []model.Comment
	if err := r.db.WithContext(ctx).Where("post_id = ?", postID).Order("lft ASC").Find(&rows).Error; err != nil {
		r.log.Error("get comment tree failed", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// GetSubtree returns comments in the subtree rooted at commentID within postID.
func (r *CommentRepository) GetSubtree(ctx context.Context, postID uint, commentID uint) ([]model.Comment, error) {
	var anchor model.Comment
	if err := r.db.WithContext(ctx).Where("post_id = ? AND id = ?", postID, commentID).First(&anchor).Error; err != nil {
		r.log.Error("get comment subtree anchor failed", zap.Error(err))
		return nil, err
	}
	var rows []model.Comment
	err := r.db.WithContext(ctx).
		Where("post_id = ? AND lft BETWEEN ? AND ?", postID, anchor.Lft, anchor.Rgt).
		Order("lft ASC").
		Find(&rows).Error
	if err != nil {
		r.log.Error("get comment subtree failed", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// GetByIDInPost returns a comment if it exists and belongs to postID.
func (r *CommentRepository) GetByIDInPost(ctx context.Context, postID uint, commentID uint) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.WithContext(ctx).Preload("Tags").Preload("Media").Where("post_id = ? AND id = ?", postID, commentID).First(&c).Error; err != nil {
		r.log.Error("find comment by id in post failed", zap.Error(err))
		return nil, err
	}
	return &c, nil
}

// UpdateContent updates comment text (nested-set indices unchanged).
func (r *CommentRepository) UpdateContent(ctx context.Context, postID uint, commentID uint, content string, actorUserID uint) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.WithContext(ctx).Where("post_id = ? AND id = ?", postID, commentID).First(&c).Error; err != nil {
		r.log.Error("find comment for update failed", zap.Error(err))
		return nil, err
	}
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ? AND id = ?", postID, commentID).Updates(map[string]interface{}{
		"content":    content,
		"updated_by": actorUserID,
		"updated_at": now,
	}).Error; err != nil {
		r.log.Error("update comment content failed", zap.Error(err))
		return nil, err
	}
	return r.GetByIDInPost(ctx, postID, commentID)
}

// DeleteSubtree soft-deletes a comment and descendants for this post and rebalances nested-set indices.
func (r *CommentRepository) DeleteSubtree(ctx context.Context, postID uint, commentID uint, deletedBy uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockCommentPost(tx, postID); err != nil {
			return err
		}
		defer unlockCommentPost(tx, postID)

		var anchor model.Comment
		q := tx.Where("post_id = ? AND id = ?", postID, commentID)
		if tx.Dialector.Name() == "mysql" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&anchor).Error; err != nil {
			return err
		}

		var ids []uint
		if err := tx.Model(&model.Comment{}).Where("post_id = ? AND lft BETWEEN ? AND ?", postID, anchor.Lft, anchor.Rgt).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return gorm.ErrRecordNotFound
		}

		var cmCount int64
		if err := tx.Table("comment_media").Where("comment_id IN ?", ids).Count(&cmCount).Error; err != nil {
			return err
		}
		if cmCount > 0 {
			return ErrCommentSubtreeHasMedia
		}

		width := anchor.Rgt - anchor.Lft + 1
		now := time.Now()
		if err := tx.Model(&model.Comment{}).Where("post_id = ? AND lft BETWEEN ? AND ?", postID, anchor.Lft, anchor.Rgt).Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
			"updated_by": deletedBy,
		}).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE comments SET rgt = rgt - ? WHERE post_id = ? AND deleted_at IS NULL AND rgt > ?", width, postID, anchor.Rgt).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE comments SET lft = lft - ? WHERE post_id = ? AND deleted_at IS NULL AND lft > ?", width, postID, anchor.Rgt).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		r.log.Error("delete comment subtree failed", zap.Error(err))
		return err
	}
	return nil
}

func (r *CommentRepository) List(ctx context.Context, req request.CommentListRequest) (CursorPage, error) {
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

	countQ := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("post_id = ?", req.PostID)
	if s := strings.TrimSpace(req.Search); s != "" {
		countQ = countQ.Where("content LIKE ?", "%"+s+"%")
	}

	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count comments", zap.Error(err))
		return CursorPage{}, err
	}

	q := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Preload("Tags").
		Preload("Media").
		Where("post_id = ?", req.PostID)
	if s := strings.TrimSpace(req.Search); s != "" {
		q = q.Where("content LIKE ?", "%"+s+"%")
	}

	var rows []model.Comment
	if err := q.Order("lft asc").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		r.log.Error("failed to list comments", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		return CursorPage{Items: []model.Comment{}, NextCursor: nil, PrevCursor: nil}, nil
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

// ListByPostID is a backward-compatible helper for older tests/callers.
func (r *CommentRepository) ListByPostID(ctx context.Context, postID uint, limit int) ([]model.Comment, error) {
	page, err := r.List(ctx, request.CommentListRequest{
		PostID:      postID,
		PageRequest: request.PageRequest{Page: 1, Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	items, ok := page.Items.([]model.Comment)
	if !ok {
		r.log.Error("failed to convert comment items", zap.Error(errors.New("failed to convert comment items")))
		return nil, errors.New("failed to convert comment items")
	}
	return items, nil
}

func (r *CommentRepository) PostExists(ctx context.Context, postID uint) (bool, error) {
	var id uint
	err := r.db.WithContext(ctx).Model(&model.Post{}).Select("id").Where("id = ?", postID).Limit(1).Scan(&id).Error
	if err != nil {
		r.log.Error("failed to check if post exists", zap.Error(err))
		return false, err
	}
	return id != 0, nil
}
