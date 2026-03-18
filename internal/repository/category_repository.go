package repository

import (
	"context"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, c *model.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepository) Update(ctx context.Context, c *model.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CategoryRepository) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

func (r *CategoryRepository) SoftDeleteByID(ctx context.Context, id uint, deletedBy uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Category{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Category{}, id).Error
	})
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uint) (*model.Category, error) {
	var c model.Category
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	var c model.Category
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&c).Error; err != nil {
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
		return false, err
	}
	return id != 0, nil
}

func (r *CategoryRepository) List(ctx context.Context, limit int) ([]model.Category, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	var rows []model.Category
	err := r.db.WithContext(ctx).Order("id asc").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepository) FindByIDs(ctx context.Context, ids []uint) ([]model.Category, error) {
	if len(ids) == 0 {
		return []model.Category{}, nil
	}
	var rows []model.Category
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

