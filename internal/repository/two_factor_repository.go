package repository

import (
	"context"
	"time"

	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TwoFactorRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewTwoFactorRepository(db *gorm.DB, log *zap.Logger) *TwoFactorRepository {
	return &TwoFactorRepository{db: db, log: log}
}

func (r *TwoFactorRepository) GetUserConfig(ctx context.Context, userID uint) (*model.UserTwoFactor, error) {
	var cfg model.UserTwoFactor
	err := r.db.WithContext(ctx).First(&cfg, "user_id = ?", userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Error("failed to get user config", zap.Error(err))
		return nil, err
	}
	return &cfg, nil
}

func (r *TwoFactorRepository) UpsertUserConfig(ctx context.Context, cfg *model.UserTwoFactor) error {
	err := r.db.WithContext(ctx).Save(cfg).Error
	if err != nil {
		r.log.Error("failed to upsert user config", zap.Error(err))
		return err
	}
	return nil
}

func (r *TwoFactorRepository) CreateChallenge(ctx context.Context, ch *model.TwoFactorChallenge) error {
	err := r.db.WithContext(ctx).Create(ch).Error
	if err != nil {
		r.log.Error("failed to create challenge", zap.Error(err))
		return err
	}
	return nil
}

func (r *TwoFactorRepository) FindValidChallenge(ctx context.Context, id string, now time.Time, maxAttempts int) (*model.TwoFactorChallenge, error) {
	var ch model.TwoFactorChallenge
	err := r.db.WithContext(ctx).
		Where("id = ? AND expires_at > ? AND consumed_at IS NULL AND attempts < ?", id, now, maxAttempts).
		First(&ch).Error
	if err != nil {
		r.log.Error("failed to find valid challenge", zap.Error(err))
		return nil, err
	}
	return &ch, nil
}

func (r *TwoFactorRepository) MarkChallengeUsed(ctx context.Context, id string, when time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&model.TwoFactorChallenge{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"consumed_at": when,
		}).Error
	if err != nil {
		r.log.Error("failed to mark challenge used", zap.Error(err))
		return err
	}
	return nil
}

func (r *TwoFactorRepository) IncrementAttempts(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).
		Model(&model.TwoFactorChallenge{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
	if err != nil {
		r.log.Error("failed to increment attempts", zap.Error(err))
		return err
	}
	return nil
}
