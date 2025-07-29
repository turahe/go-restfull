package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createMenuRolesTable20250708232040)
}

var createMenuRolesTable20250708232040 = &Migration{
	Name: "20250708232040_create_menu_entities_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS menu_entities (
				"model_id" UUID NOT NULL,
				"model_type" VARCHAR(255) NOT NULL,
				"menu_id" UUID NOT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "menu_entities_pkey" PRIMARY KEY ("created_at"),
				CONSTRAINT "menu_entities_model_id_fkey" FOREIGN KEY ("model_id") REFERENCES "models"("id") ON DELETE CASCADE,
				CONSTRAINT "menu_entities_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles"("id") ON DELETE CASCADE,
				CONSTRAINT "menu_entities_model_type_check" CHECK ("model_type" IN ('post', 'page', 'comment', 'media', 'taxonomy', 'organization', 'user', 'job', 'menu', 'menu_role', 'role', 'user_role', 'setting', 'content', 'tag', 'taxonomy')),
				CONSTRAINT "menu_entities_model_id_check" CHECK ("model_id" IS NOT NULL),
				CONSTRAINT "menu_entities_menu_id_check" CHECK ("menu_id" IS NOT NULL)
			);
		`)
		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS menu_entities;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
