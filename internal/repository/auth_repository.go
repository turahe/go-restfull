package repository

import (
	"context"
	"time"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateSession(ctx context.Context, s *model.AuthSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *AuthRepository) TouchSession(ctx context.Context, sessionID string, t time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.AuthSession{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Update("last_seen_at", t).Error
}

func (r *AuthRepository) RevokeSession(ctx context.Context, sessionID string, revokedBy *uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.AuthSession{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{"revoked_at": &now, "revoked_by": revokedBy}).Error
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
		return false, err
	}
	return id != "", nil
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, rt *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *AuthRepository) FindRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *AuthRepository) MarkRefreshTokenUsed(ctx context.Context, id uint, t time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id = ? AND used_at IS NULL AND revoked_at IS NULL", id).
		Update("used_at", &t).Error
}

func (r *AuthRepository) RevokeRefreshFamily(ctx context.Context, tokenFamily string, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_family = ? AND revoked_at IS NULL", tokenFamily).
		Updates(map[string]any{"revoked_at": &now, "revoked_reason": reason}).Error
}

func (r *AuthRepository) RevokeRefreshBySessionID(ctx context.Context, sessionID string, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{"revoked_at": &now, "revoked_reason": reason}).Error
}

func (r *AuthRepository) CreateRevokedJTI(ctx context.Context, j *model.RevokedJTI) error {
	return r.db.WithContext(ctx).Create(j).Error
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
		return false, err
	}
	return v != "", nil
}

