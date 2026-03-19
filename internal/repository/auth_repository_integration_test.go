//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"go-rest/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestAuthRepository_Integration_Session_RefreshToken_JTI(t *testing.T) {
	RunWithTx(t, func(tx *gorm.DB) {
		userRepo := NewUserRepository(tx, zap.NewNop())
		authRepo := NewAuthRepository(tx, zap.NewNop())
		ctx := context.Background()

		u := &model.User{Name: "Auth User", Email: "auth@example.com", Password: "secret"}
		require.NoError(t, userRepo.Create(ctx, u))

		sessionID := uuid.New().String()
		s := &model.AuthSession{
			ID:         sessionID,
			UserID:     u.ID,
			DeviceID:   "device-1",
			IPAddress:  "127.0.0.1",
			UserAgent:  "test",
			LastSeenAt: time.Now(),
		}
		require.NoError(t, authRepo.CreateSession(ctx, s))

		active, err := authRepo.SessionActive(ctx, sessionID)
		require.NoError(t, err)
		assert.True(t, active)

		hash := "refresh-hash-" + uuid.New().String()
		family := uuid.New().String()
		rt := &model.RefreshToken{
			SessionID:   sessionID,
			UserID:      u.ID,
			TokenHash:   hash,
			TokenFamily: family,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, authRepo.CreateRefreshToken(ctx, rt))
		assert.NotZero(t, rt.ID)

		found, err := authRepo.FindRefreshTokenByHash(ctx, hash)
		require.NoError(t, err)
		assert.Equal(t, rt.ID, found.ID)

		now := time.Now()
		require.NoError(t, authRepo.MarkRefreshTokenUsed(ctx, rt.ID, now))
		found2, err := authRepo.FindRefreshTokenByHash(ctx, hash)
		require.NoError(t, err)
		require.NotNil(t, found2.UsedAt)

		// RevokedJTI.JTI is char(36), so use raw UUID length.
		jti := uuid.New().String()
		revoked := &model.RevokedJTI{
			JTI:       jti,
			UserID:    u.ID,
			SessionID: sessionID,
			Reason:    "test",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		require.NoError(t, authRepo.CreateRevokedJTI(ctx, revoked))
		revokedCheck, err := authRepo.IsJTIRevoked(ctx, jti)
		require.NoError(t, err)
		assert.True(t, revokedCheck)
	})
}
