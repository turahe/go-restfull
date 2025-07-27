package cmd

import (
	"context"
	"fmt"
	"os"

	"webapi/internal/db/pgx"
	"webapi/internal/logger"
	"webapi/internal/seeders"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: "seed", Title: "Seed:"})
	rootCmd.AddCommand(
		seedCommand,
		seedFlushCommand,
		seedStatusCommand,
	)
}

var seedCommand = &cobra.Command{
	Use:     "seed",
	Short:   "Seed database with initial data",
	GroupID: "seed",
	Run: func(_ *cobra.Command, _ []string) {
		// Setup all the required dependencies
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		// Initiate context
		ctx := context.Background()

		// Get database connection
		dbConn := pgx.GetPgxPool()

		if dbConn == nil {
			logger.Log.Error("Database connection is nil")
			return
		}

		// Create the seeders table if it doesn't exist
		_, err := dbConn.Exec(
			ctx,
			`CREATE TABLE IF NOT EXISTS seeders (
					id SERIAL PRIMARY KEY,
					seeder VARCHAR(255) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT NOW()
				)`,
		)
		if err != nil {
			logger.Log.Error("Error creating seeders table", zap.Error(err))
			return
		}

		// Run all seeders
		logger.Log.Info("Starting database seeding...")

		// Initialize seeders
		seederManager := seeders.NewSeederManager(dbConn)

		// Run seeders in order
		seedersToRun := []seeders.Seeder{
			seeders.NewRoleSeeder(),
			seeders.NewUserSeeder(),
			seeders.NewUserRoleSeeder(),
			seeders.NewTaxonomySeeder(),
			seeders.NewTagSeeder(),
			seeders.NewMenuSeeder(),
			seeders.NewMenuRoleSeeder(),
			seeders.NewPostSeeder(),
			seeders.NewContentSeeder(),
			seeders.NewCommentSeeder(),
			seeders.NewSettingSeeder(),
		}

		for _, seeder := range seedersToRun {
			err := seederManager.RunSeeder(ctx, seeder)
			if err != nil {
				logger.Log.Error("Error running seeder",
					zap.String("seeder", seeder.GetName()),
					zap.Error(err))
				os.Exit(1)
			}
		}

		logger.Log.Info("Database seeding completed successfully")
	},
}

var seedFlushCommand = &cobra.Command{
	Use:     "seed:flush",
	Short:   "Clear all seeded data",
	GroupID: "seed",
	Run: func(_ *cobra.Command, _ []string) {
		// Setup all the required dependencies
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		// Initiate context
		ctx := context.Background()

		// Get database connection
		dbConn := pgx.GetPgxPool()

		if dbConn == nil {
			logger.Log.Error("Database connection is nil")
			return
		}

		logger.Log.Info("Clearing all seeded data...")

		// Clear all seeded data
		tables := []string{
			"comments",
			"posts",
			"contents",
			"tags",
			"taxonomies",
			"menu_roles",
			"menus",
			"user_roles",
			"users",
			"roles",
			"settings",
		}

		for _, table := range tables {
			_, err := dbConn.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
			if err != nil {
				logger.Log.Error("Error clearing table",
					zap.String("table", table),
					zap.Error(err))
			} else {
				logger.Log.Info("Cleared table", zap.String("table", table))
			}
		}

		// Clear seeders table
		_, err := dbConn.Exec(ctx, "DELETE FROM seeders")
		if err != nil {
			logger.Log.Error("Error clearing seeders table", zap.Error(err))
		} else {
			logger.Log.Info("Cleared seeders table")
		}

		logger.Log.Info("All seeded data cleared successfully")
	},
}

var seedStatusCommand = &cobra.Command{
	Use:     "seed:status",
	Short:   "Show seeding status",
	GroupID: "seed",
	Run: func(_ *cobra.Command, _ []string) {
		// Setup all the required dependencies
		setUpConfig()
		setUpLogger()
		setUpPostgres()

		// Initiate context
		ctx := context.Background()

		// Get database connection
		dbConn := pgx.GetPgxPool()

		if dbConn == nil {
			logger.Log.Error("Database connection is nil")
			return
		}

		// Check if seeders table exists
		var exists bool
		err := dbConn.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_name = 'seeders'
			)
		`).Scan(&exists)

		if err != nil {
			logger.Log.Error("Error checking seeders table", zap.Error(err))
			return
		}

		if !exists {
			logger.Log.Info("Seeders table does not exist - no seeding has been performed")
			return
		}

		// Get all applied seeders
		rows, err := dbConn.Query(ctx, `
			SELECT seeder, created_at 
			FROM seeders 
			ORDER BY created_at ASC
		`)
		if err != nil {
			logger.Log.Error("Error querying seeders", zap.Error(err))
			return
		}
		defer rows.Close()

		logger.Log.Info("Applied seeders:")
		for rows.Next() {
			var seederName string
			var createdAt string
			err := rows.Scan(&seederName, &createdAt)
			if err != nil {
				logger.Log.Error("Error scanning seeder row", zap.Error(err))
				continue
			}
			logger.Log.Info("Seeder applied",
				zap.String("seeder", seederName),
				zap.String("created_at", createdAt))
		}
	},
}
