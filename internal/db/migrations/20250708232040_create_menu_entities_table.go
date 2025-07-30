package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createMenuEntitiesTable20250708232040)
}

var createMenuEntitiesTable20250708232040 = &Migration{
	Name: "20250708232040_create_menu_entities_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS menu_entities (
				"menu_id" UUID NOT NULL,
				"entity_id" UUID NOT NULL,
				"entity_type" VARCHAR(255) NOT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "menu_entities_pkey" PRIMARY KEY ("menu_id", "entity_id"),
				CONSTRAINT "menu_entities_menu_id_fkey" FOREIGN KEY ("menu_id") REFERENCES "menus"("id") ON DELETE CASCADE,
				CONSTRAINT "menu_entities_menu_id_check" CHECK ("menu_id" IS NOT NULL),
				CONSTRAINT "menu_entities_entity_id_check" CHECK ("entity_id" IS NOT NULL)
			)
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
