package main

import (
	"github.com/turahe/go-restfull/internal/config"
	"github.com/turahe/go-restfull/internal/database"
	"github.com/turahe/go-restfull/internal/rbac"
	"github.com/turahe/go-restfull/internal/seeder"

	"github.com/spf13/cobra"
)

func newSeedRBACCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rbac",
		Short: "Seed default roles, permissions, and Casbin policies",
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

			enf, err := rbac.NewEnforcer(db.Gorm, cfg.CasbinModelPath)
			if err != nil {
				return err
			}
			return seeder.SeedDefaultRBAC(cmd.Context(), db.Gorm, enf)
		},
	}
}

