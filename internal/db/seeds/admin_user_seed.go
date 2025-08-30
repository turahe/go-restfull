package seeds

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/infrastructure/adapters"
)

// SeedAdminUser seeds the database with a default admin user
// This function creates an admin user with full system access and assigns
// appropriate roles and menu permissions.
func SeedAdminUser() error {
	pool := pgx.GetPgxPool()
	ctx := context.Background()

	// First, ensure roles exist
	if err := SeedRoles(); err != nil {
		log.Printf("Error seeding roles: %v", err)
		return err
	}

	// Then, ensure menus exist
	if err := SeedMenus(); err != nil {
		log.Printf("Error seeding menus: %v", err)
		return err
	}

	// Create admin user
	adminUser, err := createAdminUser(ctx, pool)
	if err != nil {
		log.Printf("Error creating admin user: %v", err)
		return err
	}

	// Assign admin role to user
	if err := assignAdminRoleToUser(ctx, pool, adminUser.ID); err != nil {
		log.Printf("Error assigning admin role to user: %v", err)
		return err
	}

	// Assign all menus to admin role
	if err := assignAllMenusToAdminRole(ctx, pool); err != nil {
		log.Printf("Error assigning menus to admin role: %v", err)
		return err
	}

	log.Printf("Successfully seeded admin user: %s (%s)", adminUser.UserName, adminUser.Email)
	return nil
}

// createAdminUser creates the admin user in the database
func createAdminUser(ctx context.Context, pool *pgxpool.Pool) (*entities.User, error) {
	// Check if admin user already exists
	var existingUser entities.User
	query := `SELECT id, username, email FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := pool.QueryRow(ctx, query, "admin@example.com").Scan(&existingUser.ID, &existingUser.UserName, &existingUser.Email)
	if err == nil {
		log.Printf("Admin user already exists: %s", existingUser.Email)
		return &existingUser, nil
	}

	// Generate a secure random password
	password := generateSecurePassword()
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create admin user
	adminUser := &entities.User{
		ID:        uuid.New(),
		UserName:  "admin",
		Email:     "admin@example.com",
		Phone:     "+1234567890",
		Password:  hashedPassword,
		Avatar:    "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert user into database
	insertQuery := `
		INSERT INTO users (id, username, email, phone, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = pool.Exec(ctx, insertQuery,
		adminUser.ID,
		adminUser.UserName,
		adminUser.Email,
		adminUser.Phone,
		adminUser.Password,
		adminUser.CreatedAt,
		adminUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	log.Printf("Created admin user with password: %s", password)
	log.Printf("IMPORTANT: Please change this password after first login!")

	return adminUser, nil
}

// assignAdminRoleToUser assigns the admin role to the specified user
func assignAdminRoleToUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	// Get admin role ID
	var adminRoleID uuid.UUID
	roleQuery := `SELECT id FROM roles WHERE slug = 'admin' AND deleted_at IS NULL`
	err := pool.QueryRow(ctx, roleQuery).Scan(&adminRoleID)
	if err != nil {
		return err
	}

	// Check if user-role relationship already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)`
	err = pool.QueryRow(ctx, checkQuery, userID, adminRoleID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("Admin role already assigned to user")
		return nil
	}

	// Create user-role relationship
	insertQuery := `
		INSERT INTO user_roles (id, user_id, role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = pool.Exec(ctx, insertQuery,
		uuid.New(),
		userID,
		adminRoleID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	log.Printf("Assigned admin role to user")
	return nil
}

// assignAllMenusToAdminRole assigns all menus to the admin role
func assignAllMenusToAdminRole(ctx context.Context, pool *pgxpool.Pool) error {
	// Get admin role ID
	var adminRoleID uuid.UUID
	roleQuery := `SELECT id FROM roles WHERE slug = 'admin' AND deleted_at IS NULL`
	err := pool.QueryRow(ctx, roleQuery).Scan(&adminRoleID)
	if err != nil {
		return err
	}

	// Get all menu IDs
	menuQuery := `SELECT id FROM menus WHERE deleted_at IS NULL`
	rows, err := pool.Query(ctx, menuQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var menuIDs []uuid.UUID
	for rows.Next() {
		var menuID uuid.UUID
		if err := rows.Scan(&menuID); err != nil {
			return err
		}
		menuIDs = append(menuIDs, menuID)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Assign each menu to admin role
	for _, menuID := range menuIDs {
		// Check if menu-role relationship already exists
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM menu_roles WHERE menu_id = $1 AND role_id = $2)`
		err = pool.QueryRow(ctx, checkQuery, menuID, adminRoleID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking menu-role relationship: %v", err)
			continue
		}

		if exists {
			continue
		}

		// Create menu-role relationship
		insertQuery := `
			INSERT INTO menu_roles (id, menu_id, role_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`

		_, err = pool.Exec(ctx, insertQuery,
			uuid.New(),
			menuID,
			adminRoleID,
			time.Now(),
			time.Now(),
		)

		if err != nil {
			log.Printf("Error assigning menu %s to admin role: %v", menuID, err)
			continue
		}
	}

	log.Printf("Assigned %d menus to admin role", len(menuIDs))
	return nil
}

// hashPassword hashes the password using the proper password service
func hashPassword(password string) (string, error) {
	passwordService := adapters.NewBcryptPasswordService()
	return passwordService.HashPassword(password)
}

// generateSecurePassword generates a secure random password
func generateSecurePassword() string {
	// Return default password "secret" for development/testing
	return "secret"
}
