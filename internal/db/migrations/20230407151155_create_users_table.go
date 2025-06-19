package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createUserTable)
}

var createUserTable = &Migration{
	Name: "20230407151155_create_users_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS users (
				"id" UUID NOT NULL,
				"username" VARCHAR(255) NOT NULL UNIQUE,
				"email" VARCHAR(255) NOT NULL UNIQUE,
			    "phone" VARCHAR(255) NULL UNIQUE,
			    "password" VARCHAR(255) NULL,
			    "email_verified_at" BIGINT NULL,
			    "phone_verified_at" BIGINT NULL,
			    "created_at" BIGINT NULL,
			    "updated_at" BIGINT NULL,
			    "deleted_at" BIGINT NULL,
			    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
			);

		`)

		if err != nil {
			return err
		}
		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS users;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
