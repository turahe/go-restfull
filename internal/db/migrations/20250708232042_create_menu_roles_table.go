package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createMenuRolesTable)
}

var createMenuRolesTable = &Migration{
	Name: "20250708232042_create_menu_roles_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS menu_roles (
				"id" UUID NOT NULL PRIMARY KEY,
				"menu_id" UUID NOT NULL,
				"role_id" UUID NOT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "menu_roles_menu_id_fkey" FOREIGN KEY ("menu_id") REFERENCES "menus"("id") ON DELETE CASCADE,
				CONSTRAINT "menu_roles_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles"("id") ON DELETE CASCADE,
				CONSTRAINT "menu_roles_menu_role_unique" UNIQUE ("menu_id", "role_id")
			);

			-- Create indexes for better performance
			CREATE INDEX IF NOT EXISTS "menu_roles_menu_id_idx" ON "menu_roles" ("menu_id");
			CREATE INDEX IF NOT EXISTS "menu_roles_role_id_idx" ON "menu_roles" ("role_id");
		`)

		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS menu_roles;
		`)
		if err != nil {
			return err
		}
		return nil
	},
}
