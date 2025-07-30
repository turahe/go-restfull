package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createUserRolesTable20250708232038)
}

var createUserRolesTable20250708232038 = &Migration{
	Name: "20250708232038_create_user_roles_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS user_roles (
				"id" UUID NOT NULL,
				"user_id" UUID NOT NULL,
				"role_id" UUID NOT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "user_roles_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "user_roles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE,
				CONSTRAINT "user_roles_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles"("id") ON DELETE CASCADE,
				CONSTRAINT "user_roles_user_role_unique" UNIQUE ("user_id", "role_id")
			);
		`)

		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS user_roles;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
