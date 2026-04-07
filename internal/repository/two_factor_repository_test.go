package repository

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestTwoFactorRepository_UserConfig_Upsert_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.UserTwoFactor{})
	repo := NewTwoFactorRepository(db, zap.NewNop())

	got, err := repo.GetUserConfig(ctx, 1)
	assert.NoError(t, err)
	assert.Nil(t, got)

	now := time.Now()
	cfg := &model.UserTwoFactor{UserID: 1, SecretEnc: "x", Enabled: false, CreatedAt: now, UpdatedAt: now}
	assert.NoError(t, repo.UpsertUserConfig(ctx, cfg))

	got2, err := repo.GetUserConfig(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, got2)
	assert.Equal(t, "x", got2.SecretEnc)
}

func TestTwoFactorRepository_ChallengeLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.TwoFactorChallenge{})
	repo := NewTwoFactorRepository(db, zap.NewNop())

	now := time.Now()
	ch := &model.TwoFactorChallenge{
		ID:        "c1",
		UserID:    1,
		DeviceID:  "dev1",
		ExpiresAt: now.Add(10 * time.Minute),
		Attempts:  0,
		CreatedAt: now,
	}
	assert.NoError(t, repo.CreateChallenge(ctx, ch))

	found, err := repo.FindValidChallenge(ctx, "c1", now, 5)
	assert.NoError(t, err)
	assert.Equal(t, "c1", found.ID)

	assert.NoError(t, repo.IncrementAttempts(ctx, "c1"))
	_, err = repo.FindValidChallenge(ctx, "c1", now, 1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	usedAt := now.Add(1 * time.Minute)
	assert.NoError(t, repo.MarkChallengeUsed(ctx, "c1", usedAt))
	_, err = repo.FindValidChallenge(ctx, "c1", now, 5)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
