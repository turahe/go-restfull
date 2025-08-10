package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createRoleEntitiesTable20250708232038)
}

var createRoleEntitiesTable20250708232038 = &Migration{
	Name: "20250708232038_create_role_entities_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS role_entities (
				"id" UUID NOT NULL,
				"entity_id" UUID NOT NULL,
				"entity_type" VARCHAR(255) NOT NULL,
				"role_id" UUID NOT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "role_entities_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "role_entities_entity_id_fkey" FOREIGN KEY ("entity_id") REFERENCES "users"("id") ON DELETE CASCADE,
				CONSTRAINT "role_entities_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles"("id") ON DELETE CASCADE
			);
		`)

		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS role_entities;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
