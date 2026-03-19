package main

import (
	"go-rest/internal/config"
	"go-rest/internal/database"
	"go-rest/internal/rbac"
	"go-rest/internal/seeder"

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

