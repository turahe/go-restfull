package repository

import (
	"context"
	"math"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PostRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewPostRepository(db *gorm.DB, log *zap.Logger) *PostRepository {
	return &PostRepository{db: db, log: log}
}

func (r *PostRepository) Create(ctx context.Context, p *model.Post) error {
	err := r.db.WithContext(ctx).Create(p).Error
	if err != nil {
		r.log.Error("failed to create post", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostRepository) Update(ctx context.Context, p *model.Post) error {
	err := r.db.WithContext(ctx).Save(p).Error
	if err != nil {
		r.log.Error("failed to update post", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostRepository) DeleteByID(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&model.Post{}, id).Error
	if err != nil {
		r.log.Error("failed to delete post by id", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostRepository) SoftDeleteByID(ctx context.Context, id uint, deletedBy uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
			r.log.Error("failed to update post deleted by", zap.Error(err))
			return err
		}
		return tx.Delete(&model.Post{}, id).Error
	})
	if err != nil {
		r.log.Error("failed to soft delete post by id", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostRepository) FindByID(ctx context.Context, id uint) (*model.Post, error) {
	var p model.Post
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		r.log.Error("failed to find post by id", zap.Error(err))
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) FindBySlug(ctx context.Context, slug string) (*model.Post, error) {
	var p model.Post
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&p).Error
	if err != nil {
		r.log.Error("failed to find post by slug", zap.Error(err))
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) FindBySlugWithCategory(ctx context.Context, slug string) (*model.Post, error) {
	var p model.Post
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		Preload("Media").
		Preload("User").
		Where("slug = ?", slug).
		First(&p).Error
	if err != nil {
		r.log.Error("failed to find post by slug with category", zap.Error(err))
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) SetCategory(ctx context.Context, postID uint, categoryID uint) error {
	err := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).Update("category_id", categoryID).Error
	if err != nil {
		r.log.Error("failed to set category", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostRepository) ReplaceTags(ctx context.Context, postID uint, tags []model.Tag) error {
	p := model.Post{ID: postID}
	err := r.db.WithContext(ctx).Model(&p).Association("Tags").Replace(tags)
	if err != nil {
		r.log.Error("failed to replace tags", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var id uint
	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Select("id").
		Where("slug = ?", slug).
		Limit(1).
		Scan(&id).Error
	if err != nil {
		r.log.Error("failed to check if slug exists", zap.Error(err))
		return false, err
	}
	return id != 0, nil
}

type CursorPage struct {
	Items      any
	NextCursor *uint
	PrevCursor *uint
}

func (r *PostRepository) ListCursor(ctx context.Context, req request.PostListRequest) (CursorPage, error) {
	// Normalize pagination inputs (offset/limit style).
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if req.Title != "" {
			db = db.Where("title LIKE ?", "%"+req.Title+"%")
		}
		if req.Content != "" {
			db = db.Where("content LIKE ?", "%"+req.Content+"%")
		}
		if req.CategoryID != nil && *req.CategoryID > 0 {
			db = db.Where("category_id = ?", *req.CategoryID)
		}
		return db
	}

	// Count total rows for page metadata and next/prev existence.
	var totalRows int64
	countQ := applyFilters(r.db.WithContext(ctx).Model(&model.Post{}))
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count posts", zap.Error(err))
		return CursorPage{}, err
	}

	totalPages := 0
	if totalRows > 0 {
		totalPages = int(math.Ceil(float64(totalRows) / float64(limit)))
	}

	// Fetch page items.
	var rows []model.Post
	dataQ := applyFilters(
		r.db.WithContext(ctx).
			Model(&model.Post{}).
			Preload("Media").
			Preload("Tags").
			Preload("Category").
			Preload("User"),
	)

	if err := dataQ.
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		r.log.Error("failed to list posts", zap.Error(err))
		return CursorPage{}, err
	}

	// Handler only checks for non-nil cursors to decide next/prev.
	var nextCursor *uint
	var prevCursor *uint

	if len(rows) > 0 {
		if totalPages > 0 && page < totalPages {
			tmp := rows[len(rows)-1].ID
			nextCursor = &tmp
		}
		if page > 1 && page <= totalPages {
			tmp := rows[0].ID
			prevCursor = &tmp
		}
	}

	return CursorPage{Items: rows, NextCursor: nextCursor, PrevCursor: prevCursor}, nil
}
