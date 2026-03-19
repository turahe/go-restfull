package repository

import (
	"context"
	"testing"
	"time"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestAuthRepository_SessionActive_Revoke_Touch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.AuthSession{}, &model.RefreshToken{}, &model.RevokedJTI{})
	repo := NewAuthRepository(db)

	sess := &model.AuthSession{
		ID:         "s1",
		UserID:     1,
		DeviceID:   "dev1",
		IPAddress:  "127.0.0.1",
		UserAgent:  "ua",
		LastSeenAt: time.Now(),
	}
	assert.NoError(t, repo.CreateSession(ctx, sess))

	active, err := repo.SessionActive(ctx, "s1")
	assert.NoError(t, err)
	assert.True(t, active)

	t2 := time.Now().Add(1 * time.Minute)
	assert.NoError(t, repo.TouchSession(ctx, "s1", t2))

	var fetched model.AuthSession
	assert.NoError(t, db.WithContext(ctx).First(&fetched, "id = ?", "s1").Error)
	assert.True(t, fetched.LastSeenAt.Equal(t2))

	assert.NoError(t, repo.RevokeSession(ctx, "s1", nil))
	active, err = repo.SessionActive(ctx, "s1")
	assert.NoError(t, err)
	assert.False(t, active)
}

func TestAuthRepository_RefreshToken_Find_MarkUsed_Revoke(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.AuthSession{}, &model.RefreshToken{}, &model.RevokedJTI{})
	repo := NewAuthRepository(db)

	rt := &model.RefreshToken{
		SessionID:   "s1",
		UserID:      1,
		TokenHash:   "h1",
		TokenFamily: "fam",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	assert.NoError(t, repo.CreateRefreshToken(ctx, rt))
	assert.NotZero(t, rt.ID)

	got, err := repo.FindRefreshTokenByHash(ctx, "h1")
	assert.NoError(t, err)
	assert.Equal(t, rt.ID, got.ID)

	usedAt := time.Now()
	assert.NoError(t, repo.MarkRefreshTokenUsed(ctx, rt.ID, usedAt))

	var fetched model.RefreshToken
	assert.NoError(t, db.WithContext(ctx).First(&fetched, rt.ID).Error)
	if assert.NotNil(t, fetched.UsedAt) {
		assert.True(t, fetched.UsedAt.Equal(usedAt))
	}

	// Revoke by family should set revoked fields (only where revoked_at is null)
	assert.NoError(t, repo.RevokeRefreshFamily(ctx, "fam", "reason"))
	assert.NoError(t, db.WithContext(ctx).First(&fetched, rt.ID).Error)
	assert.NotNil(t, fetched.RevokedAt)
	assert.Equal(t, "reason", fetched.RevokedReason)
}

func TestAuthRepository_RevokeRefreshBySessionID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.AuthSession{}, &model.RefreshToken{}, &model.RevokedJTI{})
	repo := NewAuthRepository(db)

	rt := &model.RefreshToken{
		SessionID:   "s1",
		UserID:      1,
		TokenHash:   "h1",
		TokenFamily: "fam",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	assert.NoError(t, repo.CreateRefreshToken(ctx, rt))

	assert.NoError(t, repo.RevokeRefreshBySessionID(ctx, "s1", "logout"))

	var fetched model.RefreshToken
	assert.NoError(t, db.WithContext(ctx).First(&fetched, rt.ID).Error)
	assert.NotNil(t, fetched.RevokedAt)
	assert.Equal(t, "logout", fetched.RevokedReason)
}

func TestAuthRepository_JTIRevocation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.AuthSession{}, &model.RefreshToken{}, &model.RevokedJTI{})
	repo := NewAuthRepository(db)

	ok, err := repo.IsJTIRevoked(ctx, "j1")
	assert.NoError(t, err)
	assert.False(t, ok)

	j := &model.RevokedJTI{
		JTI:       "j1",
		UserID:    1,
		SessionID: "s1",
		Reason:    "logout",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	assert.NoError(t, repo.CreateRevokedJTI(ctx, j))

	ok, err = repo.IsJTIRevoked(ctx, "j1")
	assert.NoError(t, err)
	assert.True(t, ok)
}

