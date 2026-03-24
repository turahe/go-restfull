//go:build integration

package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/turahe/go-restfull/internal/database"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/testutil"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var integrationDB *gorm.DB

func TestMain(m *testing.M) {
	_ = godotenv.Load()
	cfg := integrationConfig()
	if cfg.DBName == "" || cfg.DBHost == "" {
		os.Exit(0)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	if err != nil {
		os.Exit(0)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		os.Exit(1)
	}
	defer sqlDB.Close()
	if err := database.AutoMigrate(gormDB); err != nil {
		os.Exit(1)
	}
	integrationDB = gormDB
	os.Exit(m.Run())
}

func integrationConfig() struct {
	DBHost, DBPort, DBUser, DBPassword, DBName string
} {
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = get("DB_NAME", "blog_test")
	}
	return struct {
		DBHost, DBPort, DBUser, DBPassword, DBName string
	}{
		get("DB_HOST", "127.0.0.1"),
		get("DB_PORT", "3306"),
		get("DB_USER", "root"),
		os.Getenv("DB_PASSWORD"),
		dbName,
	}
}

// RunWithTx runs fn inside a transaction and rolls it back so the DB stays clean.
// Use this in every integration test so tests are isolated and leave no data.
func RunWithTx(t *testing.T, fn func(tx *gorm.DB)) {
	t.Helper()
	if integrationDB == nil {
		t.Skip("integration DB not available")
	}
	tx := integrationDB.Begin()
	defer func() {
		_ = tx.Rollback()
	}()
	fn(tx)
}

// IntegrationModels is unused but documents models migrated in integration tests.
var _ = []any{
	&model.User{}, &model.Role{}, &model.Permission{}, &model.UserRole{}, &model.RolePermission{},
	&model.AuthSession{}, &model.RefreshToken{}, &model.RevokedJTI{}, &model.ImpersonationAudit{},
	&model.UserTwoFactor{}, &model.TwoFactorChallenge{},
	&model.Category{}, &model.Tag{}, &model.Post{}, &model.Comment{}, &model.Media{}, &model.Mediable{},
}
