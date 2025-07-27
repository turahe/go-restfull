package seeders

import (
	"context"
	"time"

	"webapi/internal/infrastructure/adapters"

	"github.com/google/uuid"
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

	users := []struct {
		ID              uuid.UUID
		Username        string
		Email           string
		Phone           string
		Password        string
		EmailVerifiedAt *time.Time
		PhoneVerifiedAt *time.Time
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}{
		{
			ID:              uuid.New(),
			Username:        "superadmin",
			Email:           "superadmin@example.com",
			Phone:           "+1234567890",
			Password:        "SuperAdmin123!",
			EmailVerifiedAt: nil,
			PhoneVerifiedAt: nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			Username:        "admin",
			Email:           "admin@example.com",
			Phone:           "+1234567891",
			Password:        "Admin123!",
			EmailVerifiedAt: nil,
			PhoneVerifiedAt: nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			Username:        "editor",
			Email:           "editor@example.com",
			Phone:           "+1234567892",
			Password:        "Editor123!",
			EmailVerifiedAt: nil,
			PhoneVerifiedAt: nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			Username:        "author",
			Email:           "author@example.com",
			Phone:           "+1234567893",
			Password:        "Author123!",
			EmailVerifiedAt: nil,
			PhoneVerifiedAt: nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			Username:        "user",
			Email:           "user@example.com",
			Phone:           "+1234567894",
			Password:        "User123!",
			EmailVerifiedAt: nil,
			PhoneVerifiedAt: nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, user := range users {
		// Hash the password
		hashedPassword, err := passwordService.HashPassword(user.Password)
		if err != nil {
			return err
		}

		// Set verification times to current time
		now := time.Now()
		user.EmailVerifiedAt = &now
		user.PhoneVerifiedAt = &now

		_, err = db.Exec(ctx, `
			INSERT INTO users (id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (email) DO NOTHING
		`, user.ID, user.Username, user.Email, user.Phone, hashedPassword, user.EmailVerifiedAt, user.PhoneVerifiedAt, user.CreatedAt, user.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
