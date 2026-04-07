package repository

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openTestDB opens a unique in-memory SQLite DB per test.
// We use glebarez/sqlite (pure Go) so it works with CGO disabled.
func openTestDB(t *testing.T, migrate ...any) *gorm.DB {
	t.Helper()
	dsn := "file:" + url.QueryEscape(t.Name()) + "?mode=memory&cache=shared"

	logMode := logger.Silent
	if loggerAll := strings.TrimSpace(os.Getenv("GORM_SQL_LOG_ALL")); loggerAll != "" {
		switch strings.ToLower(loggerAll) {
		case "1", "true", "yes", "y", "on":
			logMode = logger.Info
		}
	}
	if v := strings.TrimSpace(os.Getenv("GORM_SQL_LOG_LEVEL")); v != "" {
		switch strings.ToLower(v) {
		case "silent":
			logMode = logger.Silent
		case "error":
			logMode = logger.Error
		case "warn", "warning":
			logMode = logger.Warn
		case "info":
			logMode = logger.Info
		}
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if len(migrate) > 0 {
		if err := db.AutoMigrate(migrate...); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
	return db
}
