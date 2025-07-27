package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TagSeeder seeds initial tags
type TagSeeder struct{}

// NewTagSeeder creates a new tag seeder
func NewTagSeeder() *TagSeeder {
	return &TagSeeder{}
}

// GetName returns the seeder name
func (ts *TagSeeder) GetName() string {
	return "TagSeeder"
}

// Run executes the tag seeder
func (ts *TagSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	tags := []struct {
		ID        uuid.UUID
		Name      string
		Slug      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}{
		{
			ID:        uuid.New(),
			Name:      "Go",
			Slug:      "go",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "JavaScript",
			Slug:      "javascript",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Python",
			Slug:      "python",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "React",
			Slug:      "react",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Vue.js",
			Slug:      "vuejs",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Docker",
			Slug:      "docker",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Kubernetes",
			Slug:      "kubernetes",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "API",
			Slug:      "api",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Microservices",
			Slug:      "microservices",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Database",
			Slug:      "database",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, tag := range tags {
		_, err := db.Exec(ctx, `
			INSERT INTO tags (id, name, slug, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (slug) DO NOTHING
		`, tag.ID, tag.Name, tag.Slug, tag.CreatedAt, tag.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
