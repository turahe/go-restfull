package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createRolesTable20250708232037)
}

var createRolesTable20250708232037 = &Migration{
	Name: "20250708232037_create_roles_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS roles (
				"id" UUID NOT NULL,
				"name" VARCHAR(255) NOT NULL,
				"slug" VARCHAR(255) NOT NULL UNIQUE,
				"description" TEXT NULL,
				"is_active" BOOLEAN DEFAULT TRUE,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"deleted_at" TIMESTAMP WITH TIME ZONE NULL,
				"created_by" VARCHAR(255) NULL,
				"updated_by" VARCHAR(255) NULL,
				"deleted_by" VARCHAR(255) NULL,
				CONSTRAINT "roles_pkey" PRIMARY KEY ("id")
			);
		`)

		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS roles;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
