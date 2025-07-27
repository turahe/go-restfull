package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createMenuRolesTable20250708232040)
}

var createMenuRolesTable20250708232040 = &Migration{
	Name: "20250708232040_create_menu_roles_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS menu_roles (
				"id" UUID NOT NULL,
				"menu_id" UUID NOT NULL,
				"role_id" UUID NOT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "menu_roles_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "menu_roles_menu_id_fkey" FOREIGN KEY ("menu_id") REFERENCES "menus"("id") ON DELETE CASCADE,
				CONSTRAINT "menu_roles_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles"("id") ON DELETE CASCADE,
				CONSTRAINT "menu_roles_menu_role_unique" UNIQUE ("menu_id", "role_id")
			);
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
