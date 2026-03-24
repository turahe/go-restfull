package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TagRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewTagRepository(db *gorm.DB, log *zap.Logger) *TagRepository {
	return &TagRepository{db: db, log: log}
}

func (r *TagRepository) Create(ctx context.Context, t *model.Tag) error {
	err := r.db.WithContext(ctx).Create(t).Error
	if err != nil {
		r.log.Error("failed to create tag", zap.Error(err))
		return err
	}
	return nil
}

func (r *TagRepository) Update(ctx context.Context, t *model.Tag) error {
	err := r.db.WithContext(ctx).Save(t).Error
	if err != nil {
		r.log.Error("failed to update tag", zap.Error(err))
		return err
	}
	return nil
}

func (r *TagRepository) DeleteByID(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&model.Tag{}, id).Error
	if err != nil {
		r.log.Error("failed to delete tag by id", zap.Error(err))
		return err
	}
	return nil
}

func (r *TagRepository) FindByID(ctx context.Context, id uint) (*model.Tag, error) {
	var t model.Tag
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		r.log.Error("failed to find tag by id", zap.Error(err))
		return nil, err
	}
	return &t, nil
}

func (r *TagRepository) FindBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	var t model.Tag
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error; err != nil {
		r.log.Error("failed to find tag by slug", zap.Error(err))
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
		r.log.Error("failed to check if slug exists", zap.Error(err))
		return false, err
	}
	return id != 0, nil
}

func (r *TagRepository) List(ctx context.Context, req request.TagListRequest) (CursorPage, error) {
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

	countQ := r.db.WithContext(ctx).Model(&model.Tag{})
	if req.Name != "" {
		countQ = countQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count tags", zap.Error(err))
		return CursorPage{}, err
	}

	var rows []model.Tag
	dataQ := r.db.WithContext(ctx).Order("id asc")
	if req.Name != "" {
		dataQ = dataQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	err := dataQ.Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		r.log.Error("failed to list tags", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		return CursorPage{Items: []model.Tag{}, NextCursor: nil, PrevCursor: nil}, nil
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

func (r *TagRepository) FindByIDs(ctx context.Context, ids []uint) ([]model.Tag, error) {
	if len(ids) == 0 {
		return []model.Tag{}, nil
	}
	var rows []model.Tag
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		r.log.Error("failed to find tags by ids", zap.Error(err))
		return nil, err
	}
	return rows, nil
}
