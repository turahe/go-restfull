package seeders

import (
	"context"
	"fmt"

	"webapi/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Seeder interface defines the contract for all seeders
type Seeder interface {
	GetName() string
	Run(ctx context.Context, db *pgxpool.Pool) error
}

// SeederManager manages the execution of seeders
type SeederManager struct {
	db *pgxpool.Pool
}

// NewSeederManager creates a new seeder manager
func NewSeederManager(db *pgxpool.Pool) *SeederManager {
	return &SeederManager{
		db: db,
	}
}

// RunSeeder runs a specific seeder if it hasn't been run before
func (sm *SeederManager) RunSeeder(ctx context.Context, seeder Seeder) error {
	seederName := seeder.GetName()

	// Check if seeder has already been run
	var exists bool
	err := sm.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM seeders WHERE seeder = $1
		)
	`, seederName).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking if seeder %s exists: %w", seederName, err)
	}

	if exists {
		logger.Log.Info("Seeder already applied, skipping", zap.String("seeder", seederName))
		return nil
	}

	// Run the seeder
	logger.Log.Info("Running seeder", zap.String("seeder", seederName))

	err = seeder.Run(ctx, sm.db)
	if err != nil {
		return fmt.Errorf("error running seeder %s: %w", seederName, err)
	}

	// Record that the seeder has been run
	_, err = sm.db.Exec(ctx, `
		INSERT INTO seeders (seeder) VALUES ($1)
	`, seederName)

	if err != nil {
		return fmt.Errorf("error recording seeder %s: %w", seederName, err)
	}

	logger.Log.Info("Seeder completed successfully", zap.String("seeder", seederName))
	return nil
}
