package service

import (
	"context"

	"go-rest/internal/repository"
)

// SettingsService exposes read-only public configuration for clients (SPAs, mobile apps).
type SettingsService struct {
	repo *repository.SettingRepository
}

func NewSettingsService(repo *repository.SettingRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// Public returns non-sensitive settings for the /settings API.
// If DB isn't wired (repo is nil) or there are no public rows, it returns defaults.
func (s *SettingsService) Public(ctx context.Context) (map[string]string, error) {
	if s.repo == nil {
		return nil, nil
	}

	rows, err := s.repo.ListPublic(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	m := make(map[string]string, len(rows))
	for i := range rows {
		m[rows[i].Key] = rows[i].Value
	}
	return m, nil
}
