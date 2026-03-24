package database

import (
	"testing"

	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/testutil"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:migrate_test?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	require.NoError(t, err)

	err = AutoMigrate(db)
	require.NoError(t, err)

	// Ensure core tables exist by querying one of them
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", "users").Scan(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestAutoMigrate_AllModels(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:migrate_models_test?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	require.NoError(t, err)

	err = AutoMigrate(db)
	require.NoError(t, err)

	// Create one of each model to verify schema (smoke test)
	ctx := db.Statement.Context
	require.NoError(t, db.WithContext(ctx).Create(&model.User{Name: "u", Email: "u@x.com", Password: "x"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Category{Name: "c", Slug: "c"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Setting{Key: "k", Value: "v", IsPublic: true}).Error)
}
