package repository

import (
	"context"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestSettingRepository_ListPublic_Upsert(t *testing.T) {
	t.Parallel()
	db := openTestDB(t, &model.Setting{})
	repo := NewSettingRepository(db, zap.NewNop())
	ctx := context.Background()

	require.NoError(t, repo.Upsert(ctx, "siteTitle", "Blog", true))
	require.NoError(t, repo.Upsert(ctx, "secretBanner", "hidden", false))

	public, err := repo.ListPublic(ctx)
	require.NoError(t, err)
	require.Len(t, public, 1)
	assert.Equal(t, "siteTitle", public[0].Key)
	assert.Equal(t, "Blog", public[0].Value)
	assert.True(t, public[0].IsPublic)

	all, err := repo.ListAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestSettingRepository_FindByKey_NotFound(t *testing.T) {
	t.Parallel()
	db := openTestDB(t, &model.Setting{})
	repo := NewSettingRepository(db, zap.NewNop())

	_, err := repo.FindByKey(context.Background(), "nope")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestSettingRepository_DeleteByKey(t *testing.T) {
	t.Parallel()
	db := openTestDB(t, &model.Setting{})
	repo := NewSettingRepository(db, zap.NewNop())
	ctx := context.Background()

	require.NoError(t, repo.Upsert(ctx, "k", "v", true))
	require.NoError(t, repo.DeleteByKey(ctx, "k"))

	_, err := repo.FindByKey(ctx, "k")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
