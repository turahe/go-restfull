package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewRoleRepository(db *gorm.DB, log *zap.Logger) *RoleRepository {
	return &RoleRepository{db: db, log: log}
}

func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	err := r.db.WithContext(ctx).Create(role).Error
	if err != nil {
		r.log.Error("failed to create role", zap.Error(err))
		return err
	}
	return nil
}

func (r *RoleRepository) DeleteByID(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
	if err != nil {
		r.log.Error("failed to delete role by id", zap.Error(err))
		return err
	}
	return nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		r.log.Error("failed to find role by id", zap.Error(err))
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		r.log.Error("failed to find role by name", zap.Error(err))
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) List(ctx context.Context, req request.RoleListRequest) (CursorPage, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	// Count total rows for filtered query (for cursor existence).
	countQ := r.db.WithContext(ctx).Model(&model.Role{})
	if req.Name != "" {
		countQ = countQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		return CursorPage{}, err
	}

	// Fetch current page.
	dataQ := r.db.WithContext(ctx).Order("id asc")
	if req.Name != "" {
		dataQ = dataQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	var rows []model.Role
	if err := dataQ.
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		r.log.Error("failed to list roles", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		return CursorPage{Items: []model.Role{}, NextCursor: nil, PrevCursor: nil}, nil
	}

	// Next/Prev cursors based on classic page existence.
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
