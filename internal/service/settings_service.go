package service

import (
	"context"

	"go-rest/internal/repository"
)

// PublicSettings is the exact JSON shape exposed by GET /api/v1/settings.
// It is a map of public `settings` rows (setting_key -> value).
type PublicSettings map[string]string

var fallbackPublicSettings = map[string]string{
	"siteTitle":       "Go REST Blog",
	"siteDescription": "Blog API powered by Go, Gin, and GORM.",
	"maintenanceMode": "false",
	"defaultLocale":   "en",
}

// SettingsService exposes read-only public DB-backed configuration for clients.
type SettingsService struct {
	repo *repository.SettingRepository
}

func NewSettingsService(repo *repository.SettingRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// Public returns non-sensitive settings for the /settings API.
// If DB isn't wired (repo is nil) or there are no public rows, it returns defaults.
func (s *SettingsService) Public(ctx context.Context) (PublicSettings, error) {
	if s.repo == nil {
		return fallbackPublicSettings, nil
	}

	rows, err := s.repo.ListPublic(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return fallbackPublicSettings, nil
	}

	m := make(map[string]string, len(rows))
	for i := range rows {
		m[rows[i].Key] = rows[i].Value
	}
	return m, nil
}
