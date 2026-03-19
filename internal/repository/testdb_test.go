package repository

import (
	"net/url"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// openTestDB opens a unique in-memory SQLite DB per test.
// We use glebarez/sqlite (pure Go) so it works with CGO disabled.
func openTestDB(t *testing.T, migrate ...any) *gorm.DB {
	t.Helper()
	dsn := "file:" + url.QueryEscape(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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

