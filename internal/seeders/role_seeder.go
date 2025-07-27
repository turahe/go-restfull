package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoleSeeder seeds initial roles
type RoleSeeder struct{}

// NewRoleSeeder creates a new role seeder
func NewRoleSeeder() *RoleSeeder {
	return &RoleSeeder{}
}

// GetName returns the seeder name
func (rs *RoleSeeder) GetName() string {
	return "RoleSeeder"
}

// Run executes the role seeder
func (rs *RoleSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	roles := []struct {
		ID          uuid.UUID
		Name        string
		Description string
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}{
		{
			ID:          uuid.New(),
			Name:        "super_admin",
			Description: "Super Administrator with full system access",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "admin",
			Description: "Administrator with management access",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "editor",
			Description: "Editor with content management access",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "author",
			Description: "Author with content creation access",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "user",
			Description: "Regular user with basic access",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, role := range roles {
		_, err := db.Exec(ctx, `
			INSERT INTO roles (id, name, slug, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (slug) DO NOTHING
		`, role.ID, role.Name, role.Name, role.Description, role.CreatedAt, role.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
