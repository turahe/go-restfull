package repository

import (
	"context"
	"errors"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewCommentRepository(db *gorm.DB, log *zap.Logger) *CommentRepository {
	return &CommentRepository{db: db, log: log}
}

func (r *CommentRepository) Create(ctx context.Context, cmt *model.Comment) error {
	err := r.db.WithContext(ctx).Create(cmt).Error
	if err != nil {
		r.log.Error("failed to create comment", zap.Error(err))
		return err
	}
	return nil
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

	// Base filtered query (for Count + data fetch).
	countQ := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("post_id = ?", req.PostID)
	if req.Content != "" {
		countQ = countQ.Where("content LIKE ?", "%"+req.Content+"%")
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
	if req.Content != "" {
		q = q.Where("content LIKE ?", "%"+req.Content+"%")
	}

	var rows []model.Comment
	if err := q.Order("id asc").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
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
	page, err := r.List(ctx, request.CommentListRequest{PostID: postID, Limit: limit, Page: 1})
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
