package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SettingSeeder seeds initial settings
type SettingSeeder struct{}

// NewSettingSeeder creates a new setting seeder
func NewSettingSeeder() *SettingSeeder {
	return &SettingSeeder{}
}

// GetName returns the seeder name
func (ss *SettingSeeder) GetName() string {
	return "SettingSeeder"
}

// Run executes the setting seeder
func (ss *SettingSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	settings := []struct {
		ID        uuid.UUID
		Key       string
		Value     string
		Type      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}{
		{
			ID:        uuid.New(),
			Key:       "site_name",
			Value:     "Go RESTful API",
			Type:      "string",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "site_description",
			Value:     "A modern RESTful API built with Go and Fiber",
			Type:      "string",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "site_url",
			Value:     "http://localhost:8000",
			Type:      "string",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "admin_email",
			Value:     "admin@example.com",
			Type:      "string",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "posts_per_page",
			Value:     "10",
			Type:      "integer",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "comments_enabled",
			Value:     "true",
			Type:      "boolean",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "registration_enabled",
			Value:     "true",
			Type:      "boolean",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "maintenance_mode",
			Value:     "false",
			Type:      "boolean",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "default_user_role",
			Value:     "user",
			Type:      "string",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Key:       "session_timeout",
			Value:     "3600",
			Type:      "integer",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, setting := range settings {
		_, err := db.Exec(ctx, `
			INSERT INTO settings (id, key, value, type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (key) DO NOTHING
		`, setting.ID, setting.Key, setting.Value, setting.Type, setting.CreatedAt, setting.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
