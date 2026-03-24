package repository

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewUserRepository(db *gorm.DB, log *zap.Logger) *UserRepository {
	return &UserRepository{db: db, log: log}
}

func (r *UserRepository) loadAvatar(ctx context.Context, user *model.User) (*model.Media, error) {
	var avatar model.Media
	err := r.db.WithContext(ctx).
		Model(&model.Media{}).
		Joins("INNER JOIN user_media ON media.id = user_media.media_id").
		Where("user_media.user_id = ? AND user_media.type = ?", user.ID, "avatar").
		Limit(1).
		First(&avatar).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback when a user has no avatar row in `user_media`.
			fallback := "https://ui-avatars.com/api/?name=" + user.Name
			return &model.Media{DownloadURL: fallback}, nil
		}
		return nil, err
	}
	return &avatar, nil
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	err := r.db.WithContext(ctx).Create(u).Error
	if err != nil {
		r.log.Error("failed to create user", zap.Error(err))
		return err
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Where("email = ?", email).
		First(&u).Error
	if err != nil {
		r.log.Error("failed to find user by email", zap.Error(err))
		return nil, err
	}

	avatar, err := r.loadAvatar(ctx, &u)
	if err != nil {
		r.log.Error("failed to load user avatar", zap.Error(err))
		return nil, err
	}
	u.Avatar = &avatar.DownloadURL

	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&u, id).Error
	if err != nil {
		r.log.Error("failed to find user by id", zap.Error(err))
		return nil, err
	}

	avatar, err := r.loadAvatar(ctx, &u)
	if err != nil {
		r.log.Error("failed to load user avatar", zap.Error(err))
		return nil, err
	}
	u.Avatar = &avatar.DownloadURL

	return &u, nil
}

func (r *UserRepository) List(ctx context.Context, req request.UserListRequest) (CursorPage, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Count total rows for pagination existence.
	countQ := r.db.WithContext(ctx).Model(&model.User{})
	if req.Name != "" {
		countQ = countQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Email != "" {
		countQ = countQ.Where("email = ?", req.Email)
	}
	// req.Role is currently ignored here because User model does not include role directly.
	var totalRows int64
	if err := countQ.Count(&totalRows).Error; err != nil {
		r.log.Error("failed to count users", zap.Error(err))
		return CursorPage{}, err
	}

	// Fetch page items.
	var rows []model.User
	dataQ := r.db.WithContext(ctx)

	if req.Name != "" {
		dataQ = dataQ.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Email != "" {
		dataQ = dataQ.Where("email = ?", req.Email)
	}

	if err := dataQ.
		Preload("Roles").
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		r.log.Error("failed to list users", zap.Error(err))
		return CursorPage{}, err
	}

	if len(rows) == 0 {
		return CursorPage{Items: []model.User{}, NextCursor: nil, PrevCursor: nil}, nil
	}

	var nextCursor *uint
	if int64(offset)+int64(limit) < totalRows {
		tmp := rows[len(rows)-1].ID
		nextCursor = &tmp
	}

	var prevCursor *uint
	if page > 1 {
		tmp := rows[0].ID
		prevCursor = &tmp
	}

	// Attach avatar for each returned user row.
	for i := range rows {
		avatar, err := r.loadAvatar(ctx, &rows[i])
		if err != nil {
			r.log.Error("failed to load user avatar", zap.Error(err))
			return CursorPage{}, err
		}
		rows[i].Avatar = &avatar.DownloadURL
	}

	return CursorPage{Items: rows, NextCursor: nextCursor, PrevCursor: prevCursor}, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uint, newHash string) error {
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("password", newHash).Error
	if err != nil {
		r.log.Error("failed to update password", zap.Error(err))
		return err
	}
	return nil
}

func (r *UserRepository) UpdateEmail(ctx context.Context, userID uint, newEmail string) error {
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("email", newEmail).Error
	if err != nil {
		r.log.Error("failed to update email", zap.Error(err))
		return err
	}
	return nil
}
