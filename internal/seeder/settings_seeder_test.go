package seeder

import (
	"context"
	"testing"

	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/testutil"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSeedDefaultSettings_Idempotent(t *testing.T) {
	t.Parallel()
	db, err := gorm.Open(sqlite.Open("file:seed_settings?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Setting{}))

	ctx := context.Background()
	require.NoError(t, SeedDefaultSettings(ctx, db))
	require.NoError(t, SeedDefaultSettings(ctx, db))

	var n int64
	require.NoError(t, db.Model(&model.Setting{}).Count(&n).Error)
	assert.Equal(t, int64(len(defaultSettings)), n)

	var title model.Setting
	require.NoError(t, db.Where("setting_key = ?", "siteTitle").First(&title).Error)
	assert.Equal(t, "Go REST Blog", title.Value)
	assert.True(t, title.IsPublic)

	// Existing keys are not reset on a second seed.
	require.NoError(t, db.Model(&model.Setting{}).Where("setting_key = ?", "siteTitle").Update("value", "Custom Title").Error)
	require.NoError(t, SeedDefaultSettings(ctx, db))
	require.NoError(t, db.Where("setting_key = ?", "siteTitle").First(&title).Error)
	assert.Equal(t, "Custom Title", title.Value)
}
