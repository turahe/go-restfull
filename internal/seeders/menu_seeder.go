package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MenuSeeder seeds initial menus
type MenuSeeder struct{}

// NewMenuSeeder creates a new menu seeder
func NewMenuSeeder() *MenuSeeder {
	return &MenuSeeder{}
}

// GetName returns the seeder name
func (ms *MenuSeeder) GetName() string {
	return "MenuSeeder"
}

// Run executes the menu seeder
func (ms *MenuSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	menus := []struct {
		ID          uuid.UUID
		Name        string
		Slug        string
		Description string
		ParentID    *uuid.UUID
		URL         string
		Icon        string
		Order       int
		IsActive    bool
		IsVisible   bool
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}{
		{
			ID:          uuid.New(),
			Name:        "Dashboard",
			Slug:        "dashboard",
			Description: "Main dashboard",
			ParentID:    nil,
			URL:         "/dashboard",
			Icon:        "dashboard",
			Order:       1,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Content Management",
			Slug:        "content-management",
			Description: "Content management section",
			ParentID:    nil,
			URL:         "/content",
			Icon:        "content",
			Order:       2,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Posts",
			Slug:        "posts",
			Description: "Manage posts",
			ParentID:    nil,
			URL:         "/posts",
			Icon:        "post",
			Order:       1,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Taxonomies",
			Slug:        "taxonomies",
			Description: "Manage taxonomies",
			ParentID:    nil,
			URL:         "/taxonomies",
			Icon:        "taxonomy",
			Order:       2,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Tags",
			Slug:        "tags",
			Description: "Manage tags",
			ParentID:    nil,
			URL:         "/tags",
			Icon:        "tag",
			Order:       3,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "User Management",
			Slug:        "user-management",
			Description: "User management section",
			ParentID:    nil,
			URL:         "/users",
			Icon:        "users",
			Order:       3,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Users",
			Slug:        "users",
			Description: "Manage users",
			ParentID:    nil,
			URL:         "/users",
			Icon:        "user",
			Order:       1,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Roles",
			Slug:        "roles",
			Description: "Manage roles",
			ParentID:    nil,
			URL:         "/roles",
			Icon:        "role",
			Order:       2,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "System",
			Slug:        "system",
			Description: "System settings",
			ParentID:    nil,
			URL:         "/system",
			Icon:        "settings",
			Order:       4,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Settings",
			Slug:        "settings",
			Description: "Application settings",
			ParentID:    nil,
			URL:         "/settings",
			Icon:        "setting",
			Order:       1,
			IsActive:    true,
			IsVisible:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, menu := range menus {
		_, err := db.Exec(ctx, `
			INSERT INTO menus (id, name, slug, description, parent_id, url, icon, "order", is_active, is_visible, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (slug) DO NOTHING
		`, menu.ID, menu.Name, menu.Slug, menu.Description, menu.ParentID, menu.URL, menu.Icon, menu.Order, menu.IsActive, menu.IsVisible, menu.CreatedAt, menu.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
