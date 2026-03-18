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

