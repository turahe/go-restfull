package repository

import (
	"context"
	"errors"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewCategoryRepository(db *gorm.DB, log *zap.Logger) *CategoryRepository {
	return &CategoryRepository{db: db, log: log}
}

func (r *CategoryRepository) Create(ctx context.Context, c *model.Category) error {
	err := r.db.WithContext(ctx).Create(c).Error
	if err != nil {
		r.log.Error("failed to create category", zap.Error(err))
		return err
	}
	return nil
}

func (r *CategoryRepository) Update(ctx context.Context, c *model.Category) error {
	err := r.db.WithContext(ctx).Save(c).Error
	if err != nil {
		r.log.Error("failed to update category", zap.Error(err))
		return err
	}
	return nil
}

func (r *CategoryRepository) DeleteByID(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
	if err != nil {
		r.log.Error("failed to delete category by id", zap.Error(err))
		return err
	}
	return nil
}

func (r *CategoryRepository) SoftDeleteByID(ctx context.Context, id uint, deletedBy uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Category{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
			r.log.Error("failed to update category deleted by", zap.Error(err))
			return err
		}
		return tx.Delete(&model.Category{}, id).Error
	})
	if err != nil {
		r.log.Error("failed to soft delete category by id", zap.Error(err))
		return err
	}
	return nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uint) (*model.Category, error) {
	var c model.Category
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		r.log.Error("failed to find category by id", zap.Error(err))
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	var c model.Category
	if err := r.db.WithContext(ctx).Preload("Media").Where("slug = ?", slug).First(&c).Error; err != nil {
		r.log.Error("failed to find category by slug", zap.Error(err))
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var id uint
	err := r.db.WithContext(ctx).
		Model(&model.Category{}).
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

func (r *CategoryRepository) List(ctx context.Context, req request.CategoryListRequest) (CursorPage, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	// Count total rows for filtered query.
	countQ := r.db.WithContext(ctx).Model(&model.Category{})
	if req.Name != "" {
		countQ = countQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count categories", zap.Error(err))
		return CursorPage{}, err
	}

	// Fetch page items.
	var rows []model.Category
	dataQ := r.db.WithContext(ctx).Preload("Media").Order("id asc")
	if req.Name != "" {
		dataQ = dataQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if err := dataQ.Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		r.log.Error("failed to list categories", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		r.log.Error("no categories found", zap.Error(errors.New("no categories found")))
		return CursorPage{Items: []model.Category{}, NextCursor: nil, PrevCursor: nil}, nil
	}

	var nextCursor *uint
	var prevCursor *uint

	// Next exists if there are more rows beyond this page window.
	if int64(offset)+int64(limit) < totalRows {
		tmp := rows[len(rows)-1].ID
		nextCursor = &tmp
	}
	// Prev exists if offset > 0 (page > 1) and we actually have data for this page.
	if offset > 0 {
		tmp := rows[0].ID
		prevCursor = &tmp
	}

	return CursorPage{Items: rows, NextCursor: nextCursor, PrevCursor: prevCursor}, nil
}

func (r *CategoryRepository) FindByIDs(ctx context.Context, ids []uint) ([]model.Category, error) {
	if len(ids) == 0 {
		r.log.Error("no categories found", zap.Error(errors.New("no categories found")))
		return []model.Category{}, errors.New("no categories found")
	}
	var rows []model.Category
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	if err != nil {
		r.log.Error("failed to find categories by ids", zap.Error(err))
		return nil, err
	}
	return rows, nil
}
