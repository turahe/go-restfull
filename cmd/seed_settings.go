package main

import (
	"github.com/turahe/go-restfull/internal/config"
	"github.com/turahe/go-restfull/internal/database"
	"github.com/turahe/go-restfull/internal/seeder"

	"github.com/spf13/cobra"
)

func newSeedSettingsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "settings",
		Short: "Seed default application settings (insert missing keys only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			db, err := database.ConnectMySQL(cfg, nil)
			if err != nil {
				return err
			}
			defer func() { _ = db.SQL.Close() }()

			if err := database.AutoMigrate(db.Gorm); err != nil {
				return err
			}
			return seeder.SeedDefaultSettings(cmd.Context(), db.Gorm)
		},
	}
}
