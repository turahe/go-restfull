package seeds

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/infrastructure/adapters"
)

// SeedMenus seeds the database with default menu items
// This function creates menus using the menu repository with nested set operations
func SeedMenus() error {
	pool := pgx.GetPgxPool()
	ctx := context.Background()

	// Create menu repository instance
	menuRepo := adapters.NewPostgresMenuRepository(pool, nil) // nil for Redis client in seeder

	// Define menu structure with parent-child relationships
	menuData := []struct {
		name        string
		slug        string
		description string
		url         string
		icon        string
		parentSlug  *string
	}{
		// Root level menus
		{"Dashboard", "dashboard", "Main dashboard", "/dashboard", "dashboard", nil},
		{"Permissions", "permissions", "User permissions and access control", "/permissions", "shield", nil},
		{"Blog", "blog", "Blog and content management", "/blog", "file-text", nil},
		{"Settings", "settings", "System settings and configuration", "/settings", "settings", nil},

		// Children of Permissions
		{"Users", "users", "User management", "/permissions/users", "users", stringPtr("permissions")},
		{"Roles", "roles", "Role management", "/permissions/roles", "shield", stringPtr("permissions")},
		{"Menus", "menus", "Menu management", "/permissions/menus", "menu", stringPtr("permissions")},

		// Children of Blog
		{"Posts", "posts", "Post management", "/blog/posts", "file-text", stringPtr("blog")},
		{"Taxonomy", "taxonomy", "Taxonomy and categories", "/blog/taxonomy", "tags", stringPtr("blog")},
		{"Tags", "tags", "Tag management", "/blog/tags", "tag", stringPtr("blog")},

		// Children of Settings
		{"Updates", "updates", "System updates", "/settings/updates", "refresh-cw", stringPtr("settings")},
		{"Appearance", "appearance", "System appearance and themes", "/settings/appearance", "palette", stringPtr("settings")},
	}

	// Track created menus for parent-child relationships
	createdMenus := make(map[string]*entities.Menu)

	// Create menus in order (parents first, then children)
	for _, menuItem := range menuData {
		// Check if menu already exists
		exists, err := menuRepo.ExistsBySlug(ctx, menuItem.slug)
		if err != nil {
			log.Printf("Error checking menu existence for %s: %v", menuItem.slug, err)
			continue
		}
		if exists {
			log.Printf("Menu %s already exists, skipping", menuItem.slug)
			continue
		}

		// Create menu entity
		menu := &entities.Menu{
			ID:          uuid.New(),
			Name:        menuItem.name,
			Slug:        menuItem.slug,
			Description: menuItem.description,
			URL:         menuItem.url,
			Icon:        menuItem.icon,
			IsActive:    true,
			IsVisible:   true,
			Target:      "_self",
			CreatedBy:   uuid.Nil, // Will be handled by repository fallback
			UpdatedBy:   uuid.Nil, // Will be handled by repository fallback
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Set parent ID if this is a child menu
		if menuItem.parentSlug != nil {
			if parentMenu, exists := createdMenus[*menuItem.parentSlug]; exists {
				menu.ParentID = &parentMenu.ID
			}
		}

		// Create menu using repository
		if err := menuRepo.Create(ctx, menu); err != nil {
			log.Printf("Error creating menu %s: %v", menuItem.slug, err)
			continue
		}

		// Store created menu for parent-child relationships
		createdMenus[menuItem.slug] = menu
		log.Printf("Successfully created menu: %s (%s)", menu.Name, menu.Slug)
	}

	log.Printf("Menu seeding completed. Created %d menus", len(createdMenus))
	return nil
}
