package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createMenusTable20250708232039)
}

var createMenusTable20250708232039 = &Migration{
	Name: "20250708232039_create_menus_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS menus (
				"id" UUID NOT NULL,
				"name" VARCHAR(255) NOT NULL,
				"slug" VARCHAR(255) NOT NULL UNIQUE,
				"description" TEXT NULL,
				"url" VARCHAR(500) NULL,
				"icon" VARCHAR(100) NULL,
				"parent_id" UUID NULL,
				"record_left" BIGINT NOT NULL DEFAULT 0,
				"record_right" BIGINT NOT NULL DEFAULT 0,
				"record_ordering" BIGINT NOT NULL DEFAULT 0,
				"is_active" BOOLEAN DEFAULT TRUE,
				"is_visible" BOOLEAN DEFAULT TRUE,
				"target" VARCHAR(50) DEFAULT '_self',
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"deleted_at" TIMESTAMP WITH TIME ZONE NULL,
				"created_by" VARCHAR(255) NULL,
				"updated_by" VARCHAR(255) NULL,
				"deleted_by" VARCHAR(255) NULL,
				CONSTRAINT "menus_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "menus_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "menus"("id") ON DELETE SET NULL,
				CONSTRAINT "menus_record_left_check" CHECK ("record_left" >= 0),
				CONSTRAINT "menus_record_right_check" CHECK ("record_right" >= 0),
				CONSTRAINT "menus_record_ordering_check" CHECK ("record_ordering" >= 0),
				CONSTRAINT "menus_record_left_right_check" CHECK ("record_right" > "record_left")
			);
		`)
		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS menus;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
