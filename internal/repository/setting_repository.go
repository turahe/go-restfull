package repository

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SettingRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSettingRepository(db *gorm.DB, log *zap.Logger) *SettingRepository {
	if log == nil {
		log = zap.NewNop()
	}
	return &SettingRepository{db: db, log: log}
}

// ListPublic returns all settings that may be exposed without authentication.
func (r *SettingRepository) ListPublic(ctx context.Context) ([]model.Setting, error) {
	var rows []model.Setting
	if err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Order("setting_key").
		Find(&rows).Error; err != nil {
		r.log.Error("failed to list public settings", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// ListAll returns every setting row (for admin use).
func (r *SettingRepository) ListAll(ctx context.Context) ([]model.Setting, error) {
	var rows []model.Setting
	if err := r.db.WithContext(ctx).Order("setting_key").Find(&rows).Error; err != nil {
		r.log.Error("failed to list settings", zap.Error(err))
		return nil, err
	}
	return rows, nil
}

// FindByKey loads one setting by its logical key (setting_key column).
func (r *SettingRepository) FindByKey(ctx context.Context, key string) (*model.Setting, error) {
	var s model.Setting
	if err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.log.Error("failed to find setting by key", zap.String("key", key), zap.Error(err))
		return nil, err
	}
	return &s, nil
}

// Upsert creates or updates a setting row by key.
func (r *SettingRepository) Upsert(ctx context.Context, key, value string, isPublic bool) error {
	var s model.Setting
	err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Use a map so is_public=false is persisted (GORM skips zero-value struct fields on Create).
		row := map[string]any{
			"setting_key": key,
			"value":       value,
			"is_public":   isPublic,
		}
		if err := r.db.WithContext(ctx).Table((&model.Setting{}).TableName()).Create(row).Error; err != nil {
			r.log.Error("failed to create setting", zap.String("key", key), zap.Error(err))
			return err
		}
		return nil
	}
	if err != nil {
		r.log.Error("failed to load setting for upsert", zap.String("key", key), zap.Error(err))
		return err
	}
	s.Value = value
	s.IsPublic = isPublic
	if err := r.db.WithContext(ctx).Save(&s).Error; err != nil {
		r.log.Error("failed to update setting", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// DeleteByKey removes a setting row.
func (r *SettingRepository) DeleteByKey(ctx context.Context, key string) error {
	if err := r.db.WithContext(ctx).Where("setting_key = ?", key).Delete(&model.Setting{}).Error; err != nil {
		r.log.Error("failed to delete setting", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}
