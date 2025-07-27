package seeders

import (
	"context"

	"webapi/internal/domain/entities"
	"webapi/internal/infrastructure/adapters"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserSeeder seeds initial users
type UserSeeder struct{}

// NewUserSeeder creates a new user seeder
func NewUserSeeder() *UserSeeder {
	return &UserSeeder{}
}

// GetName returns the seeder name
func (us *UserSeeder) GetName() string {
	return "UserSeeder"
}

// Run executes the user seeder
func (us *UserSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	// Create password service for hashing passwords
	passwordService := adapters.NewBcryptPasswordService()

	// Create user repository using the adapter
	userRepository := adapters.NewPostgresUserRepository(db, nil) // nil for redis client in seeder context

	// Define users using the proper User entity
	userData := []struct {
		username string
		email    string
		phone    string
		password string
	}{
		{"superadmin", "superadmin@example.com", "+1234567890", "SuperAdmin123!"},
		{"admin", "admin@example.com", "+1234567891", "Admin123!"},
		{"editor", "editor@example.com", "+1234567892", "Editor123!"},
		{"author", "author@example.com", "+1234567893", "Author123!"},
		{"user", "user@example.com", "+1234567894", "User123!"},
	}

	for _, data := range userData {
		// Create user entity using the domain constructor
		user, err := entities.NewUser(data.username, data.email, data.phone, data.password)
		if err != nil {
			return err
		}

		// Hash the password using the domain service
		hashedPassword, err := passwordService.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword

		// Verify email and phone for seeded users
		user.VerifyEmail()
		user.VerifyPhone()

		// Use the repository to create the user
		// This follows the same pattern as the rest of the application
		err = userRepository.Create(ctx, user)
		if err != nil {
			// Check if it's a duplicate email error (which is expected)
			// In a real implementation, you might want to handle this differently
			return err
		}
	}

	return nil
}
