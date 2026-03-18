package repository

import (
	"context"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, t *model.Tag) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TagRepository) Update(ctx context.Context, t *model.Tag) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *TagRepository) DeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Tag{}, id).Error
}

func (r *TagRepository) FindByID(ctx context.Context, id uint) (*model.Tag, error) {
	var t model.Tag
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TagRepository) FindBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	var t model.Tag
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TagRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var id uint
	err := r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Select("id").
		Where("slug = ?", slug).
		Limit(1).
		Scan(&id).Error
	if err != nil {
		return false, err
	}
	return id != 0, nil
}

func (r *TagRepository) List(ctx context.Context, limit int) ([]model.Tag, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}
	var rows []model.Tag
	err := r.db.WithContext(ctx).Order("id asc").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *TagRepository) FindByIDs(ctx context.Context, ids []uint) ([]model.Tag, error) {
	if len(ids) == 0 {
		return []model.Tag{}, nil
	}
	var rows []model.Tag
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

