package seeders

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRoleSeeder seeds user-role relationships
type UserRoleSeeder struct{}

// NewUserRoleSeeder creates a new user role seeder
func NewUserRoleSeeder() *UserRoleSeeder {
	return &UserRoleSeeder{}
}

// GetName returns the seeder name
func (urs *UserRoleSeeder) GetName() string {
	return "UserRoleSeeder"
}

// Run executes the user role seeder
func (urs *UserRoleSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	// Get user IDs
	var superAdminID, adminID, editorID, authorID, userID uuid.UUID

	err := db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'superadmin'`).Scan(&superAdminID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'admin'`).Scan(&adminID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'editor'`).Scan(&editorID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'author'`).Scan(&authorID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'user'`).Scan(&userID)
	if err != nil {
		return err
	}

	// Get role IDs
	var superAdminRoleID, adminRoleID, editorRoleID, authorRoleID, userRoleID uuid.UUID

	err = db.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'super_admin'`).Scan(&superAdminRoleID)
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

	// Assign roles to users
	userRoles := []struct {
		ID     uuid.UUID
		UserID uuid.UUID
		RoleID uuid.UUID
	}{
		{ID: uuid.New(), UserID: superAdminID, RoleID: superAdminRoleID},
		{ID: uuid.New(), UserID: adminID, RoleID: adminRoleID},
		{ID: uuid.New(), UserID: editorID, RoleID: editorRoleID},
		{ID: uuid.New(), UserID: authorID, RoleID: authorRoleID},
		{ID: uuid.New(), UserID: userID, RoleID: userRoleID},
	}

	for _, userRole := range userRoles {
		_, err := db.Exec(ctx, `
			INSERT INTO user_roles (id, user_id, role_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, userRole.ID, userRole.UserID, userRole.RoleID)

		if err != nil {
			return err
		}
	}

	return nil
}
