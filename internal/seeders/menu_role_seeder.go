package seeders

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MenuRoleSeeder seeds menu-role relationships
type MenuRoleSeeder struct{}

// NewMenuRoleSeeder creates a new menu role seeder
func NewMenuRoleSeeder() *MenuRoleSeeder {
	return &MenuRoleSeeder{}
}

// GetName returns the seeder name
func (mrs *MenuRoleSeeder) GetName() string {
	return "MenuRoleSeeder"
}

// Run executes the menu role seeder
func (mrs *MenuRoleSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	// Get role IDs
	var superAdminRoleID, adminRoleID, editorRoleID, authorRoleID, userRoleID uuid.UUID

	err := db.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'super_admin'`).Scan(&superAdminRoleID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'admin'`).Scan(&adminRoleID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'editor'`).Scan(&editorRoleID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'author'`).Scan(&authorRoleID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'user'`).Scan(&userRoleID)
	if err != nil {
		return err
	}

	// Get menu IDs
	var dashboardMenuID, contentMenuID, postsMenuID, taxonomiesMenuID, tagsMenuID, userMenuID, usersMenuID, rolesMenuID, systemMenuID, settingsMenuID uuid.UUID

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'dashboard'`).Scan(&dashboardMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'content-management'`).Scan(&contentMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'posts'`).Scan(&postsMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'taxonomies'`).Scan(&taxonomiesMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'tags'`).Scan(&tagsMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'user-management'`).Scan(&userMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'users'`).Scan(&usersMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'roles'`).Scan(&rolesMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'system'`).Scan(&systemMenuID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM menus WHERE slug = 'settings'`).Scan(&settingsMenuID)
	if err != nil {
		return err
	}

	// Assign roles to menus (super admin gets access to everything)
	menuRoles := []struct {
		MenuID uuid.UUID
		RoleID uuid.UUID
	}{
		// Super admin gets access to everything
		{MenuID: dashboardMenuID, RoleID: superAdminRoleID},
		{MenuID: contentMenuID, RoleID: superAdminRoleID},
		{MenuID: postsMenuID, RoleID: superAdminRoleID},
		{MenuID: taxonomiesMenuID, RoleID: superAdminRoleID},
		{MenuID: tagsMenuID, RoleID: superAdminRoleID},
		{MenuID: userMenuID, RoleID: superAdminRoleID},
		{MenuID: usersMenuID, RoleID: superAdminRoleID},
		{MenuID: rolesMenuID, RoleID: superAdminRoleID},
		{MenuID: systemMenuID, RoleID: superAdminRoleID},
		{MenuID: settingsMenuID, RoleID: superAdminRoleID},

		// Admin gets access to most things
		{MenuID: dashboardMenuID, RoleID: adminRoleID},
		{MenuID: contentMenuID, RoleID: adminRoleID},
		{MenuID: postsMenuID, RoleID: adminRoleID},
		{MenuID: taxonomiesMenuID, RoleID: adminRoleID},
		{MenuID: tagsMenuID, RoleID: adminRoleID},
		{MenuID: userMenuID, RoleID: adminRoleID},
		{MenuID: usersMenuID, RoleID: adminRoleID},
		{MenuID: systemMenuID, RoleID: adminRoleID},
		{MenuID: settingsMenuID, RoleID: adminRoleID},

		// Editor gets access to content management
		{MenuID: dashboardMenuID, RoleID: editorRoleID},
		{MenuID: contentMenuID, RoleID: editorRoleID},
		{MenuID: postsMenuID, RoleID: editorRoleID},
		{MenuID: taxonomiesMenuID, RoleID: editorRoleID},
		{MenuID: tagsMenuID, RoleID: editorRoleID},

		// Author gets access to posts and tags
		{MenuID: dashboardMenuID, RoleID: authorRoleID},
		{MenuID: postsMenuID, RoleID: authorRoleID},
		{MenuID: tagsMenuID, RoleID: authorRoleID},

		// User gets access to dashboard only
		{MenuID: dashboardMenuID, RoleID: userRoleID},
	}

	for _, menuRole := range menuRoles {
		_, err := db.Exec(ctx, `
			INSERT INTO menu_roles (menu_id, role_id)
			VALUES ($1, $2)
			ON CONFLICT (menu_id, role_id) DO NOTHING
		`, menuRole.MenuID, menuRole.RoleID)

		if err != nil {
			return err
		}
	}

	return nil
}
