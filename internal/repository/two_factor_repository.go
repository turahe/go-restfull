package repository

import (
	"context"
	"time"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type TwoFactorRepository struct {
	db *gorm.DB
}

func NewTwoFactorRepository(db *gorm.DB) *TwoFactorRepository {
	return &TwoFactorRepository{db: db}
}

func (r *TwoFactorRepository) GetUserConfig(ctx context.Context, userID uint) (*model.UserTwoFactor, error) {
	var cfg model.UserTwoFactor
	err := r.db.WithContext(ctx).First(&cfg, "user_id = ?", userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *TwoFactorRepository) UpsertUserConfig(ctx context.Context, cfg *model.UserTwoFactor) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *TwoFactorRepository) CreateChallenge(ctx context.Context, ch *model.TwoFactorChallenge) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

func (r *TwoFactorRepository) FindValidChallenge(ctx context.Context, id string, now time.Time, maxAttempts int) (*model.TwoFactorChallenge, error) {
	var ch model.TwoFactorChallenge
	err := r.db.WithContext(ctx).
		Where("id = ? AND expires_at > ? AND consumed_at IS NULL AND attempts < ?", id, now, maxAttempts).
		First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *TwoFactorRepository) MarkChallengeUsed(ctx context.Context, id string, when time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.TwoFactorChallenge{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"consumed_at": when,
		}).Error
}

func (r *TwoFactorRepository) IncrementAttempts(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.TwoFactorChallenge{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
}

