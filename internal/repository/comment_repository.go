package repository

import (
	"context"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, cmt *model.Comment) error {
	return r.db.WithContext(ctx).Create(cmt).Error
}

func (r *CommentRepository) ReplaceTags(ctx context.Context, commentID uint, tags []model.Tag) error {
	c := model.Comment{ID: commentID}
	return r.db.WithContext(ctx).Model(&c).Association("Tags").Replace(tags)
}

func (r *CommentRepository) ListByPostID(ctx context.Context, postID uint, limit int) ([]model.Comment, error) {
	var rows []model.Comment
	err := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Preload("Tags").
		Where("post_id = ?", postID).
		Order("id asc").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CommentRepository) PostExists(ctx context.Context, postID uint) (bool, error) {
	var id uint
	err := r.db.WithContext(ctx).Model(&model.Post{}).Select("id").Where("id = ?", postID).Limit(1).Scan(&id).Error
	if err != nil {
		return false, err
	}
	return id != 0, nil
}

