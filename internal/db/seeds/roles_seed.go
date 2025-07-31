package seeds

import (
	"context"
	"log"

	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/infrastructure/adapters"
)

// SeedRoles seeds the database with default roles
func SeedRoles() error {
	pool := pgx.GetPgxPool()
	roleRepo := adapters.NewPostgresRoleRepository(pool, nil)

	ctx := context.Background()

	// Default roles to create
	defaultRoles := []struct {
		name        string
		slug        string
		description string
	}{
		{"Admin", "admin", "Administrator with full system access"},
		{"User", "user", "Default user role with basic access"},
		{"Moderator", "moderator", "Moderator with content management access"},
		{"Editor", "editor", "Editor with content creation and editing access"},
		{"Viewer", "viewer", "Viewer with read-only access"},
	}

	for _, roleData := range defaultRoles {
		// Check if role already exists
		exists, err := roleRepo.ExistsBySlug(ctx, roleData.slug)
		if err != nil {
			log.Printf("Error checking if role %s exists: %v", roleData.slug, err)
			continue
		}

		if exists {
			log.Printf("Role %s already exists, skipping", roleData.slug)
			continue
		}

		// Create role
		role, err := entities.NewRole(roleData.name, roleData.slug, roleData.description)
		if err != nil {
			log.Printf("Error creating role %s: %v", roleData.slug, err)
			continue
		}

		err = roleRepo.Create(ctx, role)
		if err != nil {
			log.Printf("Error saving role %s: %v", roleData.slug, err)
			continue
		}

		log.Printf("Successfully created role: %s (%s)", roleData.name, roleData.slug)
	}

	return nil
}

// EnsureDefaultUserRole ensures the default "user" role exists
func EnsureDefaultUserRole() error {
	pool := pgx.GetPgxPool()
	roleRepo := adapters.NewPostgresRoleRepository(pool, nil)

	ctx := context.Background()

	// Check if default user role exists
	exists, err := roleRepo.ExistsBySlug(ctx, "user")
	if err != nil {
		return err
	}

	if !exists {
		// Create default user role
		role, err := entities.NewRole("User", "user", "Default user role with basic access")
		if err != nil {
			return err
		}

		err = roleRepo.Create(ctx, role)
		if err != nil {
			return err
		}

		log.Printf("Created default user role")
	}

	return nil
}
