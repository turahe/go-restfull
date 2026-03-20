package seeder

import (
	"context"
	"errors"

	"go-rest/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// defaultSettings is the baseline key/value set. SeedDefaultSettings inserts any key
// that is not yet present and does not overwrite existing rows (safe to re-run).
var defaultSettings = []struct {
	Key      string
	Value    string
	IsPublic bool
}{
	{Key: "siteTitle", Value: "Go REST Blog", IsPublic: true},
	{Key: "siteDescription", Value: "Blog API powered by Go, Gin, and GORM.", IsPublic: true},
	{Key: "maintenanceMode", Value: "false", IsPublic: true},
	{Key: "defaultLocale", Value: "en", IsPublic: true},
}

// SeedDefaultSettings ensures baseline `settings` rows exist (public site metadata and flags).
func SeedDefaultSettings(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("db is required")
	}
	repo := repository.NewSettingRepository(db, zap.NewNop())
	for _, row := range defaultSettings {
		if _, err := repo.FindByKey(ctx, row.Key); err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := repo.Upsert(ctx, row.Key, row.Value, row.IsPublic); err != nil {
			return err
		}
	}
	return nil
}
