package repository

import (
	"context"
	"time"

	"github.com/turahe/go-restfull/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewAuthRepository(db *gorm.DB, log *zap.Logger) *AuthRepository {
	return &AuthRepository{db: db, log: log}
}

func (r *AuthRepository) CreateSession(ctx context.Context, s *model.AuthSession) error {
	err := r.db.WithContext(ctx).Create(s).Error
	if err != nil {
		r.log.Error("failed to create session", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) TouchSession(ctx context.Context, sessionID string, t time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&model.AuthSession{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Update("last_seen_at", t).Error
	if err != nil {
		r.log.Error("failed to touch session", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) RevokeSession(ctx context.Context, sessionID string, revokedBy *uint) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&model.AuthSession{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{"revoked_at": &now, "revoked_by": revokedBy}).Error
	if err != nil {
		r.log.Error("failed to revoke session", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) SessionActive(ctx context.Context, sessionID string) (bool, error) {
	var id string
	err := r.db.WithContext(ctx).
		Model(&model.AuthSession{}).
		Select("id").
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Limit(1).
		Scan(&id).Error
	if err != nil {
		r.log.Error("failed to check if session is active", zap.Error(err))
		return false, err
	}
	return id != "", nil
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, rt *model.RefreshToken) error {
	err := r.db.WithContext(ctx).Create(rt).Error
	if err != nil {
		r.log.Error("failed to create refresh token", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) FindRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&rt).Error
	if err != nil {
		r.log.Error("failed to find refresh token by hash", zap.Error(err))
		return nil, err
	}
	return &rt, nil
}

func (r *AuthRepository) MarkRefreshTokenUsed(ctx context.Context, id uint, t time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id = ? AND used_at IS NULL AND revoked_at IS NULL", id).
		Update("used_at", &t).Error
	if err != nil {
		r.log.Error("failed to mark refresh token used", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) RevokeRefreshFamily(ctx context.Context, tokenFamily string, reason string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_family = ? AND revoked_at IS NULL", tokenFamily).
		Updates(map[string]any{"revoked_at": &now, "revoked_reason": reason}).Error
	if err != nil {
		r.log.Error("failed to revoke refresh family", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) RevokeRefreshBySessionID(ctx context.Context, sessionID string, reason string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{"revoked_at": &now, "revoked_reason": reason}).Error
	if err != nil {
		r.log.Error("failed to revoke refresh by session id", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) CreateRevokedJTI(ctx context.Context, j *model.RevokedJTI) error {
	err := r.db.WithContext(ctx).Create(j).Error
	if err != nil {
		r.log.Error("failed to create revoked jti", zap.Error(err))
		return err
	}
	return nil
}

func (r *AuthRepository) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	var v string
	err := r.db.WithContext(ctx).
		Model(&model.RevokedJTI{}).
		Select("jti").
		Where("jti = ? AND expires_at > ?", jti, time.Now()).
		Limit(1).
		Scan(&v).Error
	if err != nil {
		r.log.Error("failed to check if jti is revoked", zap.Error(err))
		return false, err
	}
	return v != "", nil
}
