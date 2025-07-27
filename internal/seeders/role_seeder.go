package seeders

import (
	"context"

	"webapi/internal/domain/entities"
	"webapi/internal/infrastructure/adapters"

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
	// Create role repository using the adapter
	roleRepository := adapters.NewPostgresRoleRepository(db, nil) // nil for redis client in seeder context

	// Define roles using the proper Role entity
	roleData := []struct {
		name        string
		slug        string
		description string
	}{
		{"Super Administrator", "super_admin", "Super Administrator with full system access"},
		{"Administrator", "admin", "Administrator with management access"},
		{"Editor", "editor", "Editor with content management access"},
		{"Author", "author", "Author with content creation access"},
		{"User", "user", "Regular user with basic access"},
	}

	for _, data := range roleData {
		// Create role entity using the domain constructor
		role, err := entities.NewRole(data.name, data.slug, data.description)
		if err != nil {
			return err
		}

		// Use the repository to create the role
		// This follows the same pattern as the rest of the application
		err = roleRepository.Create(ctx, role)
		if err != nil {
			// Check if it's a duplicate slug error (which is expected)
			// In a real implementation, you might want to handle this differently
			return err
		}
	}

	return nil
}
