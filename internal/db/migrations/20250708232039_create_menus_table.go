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
		// Create the menus table
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS menus (
				"id" UUID NOT NULL,
				"name" VARCHAR(255) NOT NULL,
				"slug" VARCHAR(255) NOT NULL UNIQUE,
				"description" TEXT NULL,
				"url" VARCHAR(500) NULL,
				"icon" VARCHAR(100) NULL,
				"parent_id" UUID NULL,
				"record_left" BIGINT NOT NULL DEFAULT NULL,
				"record_right" BIGINT NOT NULL DEFAULT NULL,
				"record_ordering" BIGINT NOT NULL DEFAULT NULL,
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
				CONSTRAINT "menus_record_left_right_check" CHECK ("record_right" > "record_left"),
				CONSTRAINT "menus_is_active_check" CHECK ("is_active" IN (TRUE, FALSE)),
				CONSTRAINT "menus_is_visible_check" CHECK ("is_visible" IN (TRUE, FALSE)),
				CONSTRAINT "menus_target_check" CHECK ("target" IN ('_self', '_blank', '_parent', '_top')),
				CONSTRAINT "menus_created_by_check" CHECK ("created_by" IS NOT NULL),
				CONSTRAINT "menus_updated_by_check" CHECK ("updated_by" IS NOT NULL),
				CONSTRAINT "menus_deleted_by_check" CHECK ("deleted_by" IS NOT NULL)
			)
		`)
		if err != nil {
			return err
		}

		// Create indexes for nested set operations
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE INDEX IF NOT EXISTS "menus_record_left_idx" ON "menus" ("record_left");
			CREATE INDEX IF NOT EXISTS "menus_record_right_idx" ON "menus" ("record_right");
			CREATE INDEX IF NOT EXISTS "menus_record_ordering_idx" ON "menus" ("record_ordering");
			CREATE INDEX IF NOT EXISTS "menus_parent_id_idx" ON "menus" ("parent_id");
			CREATE INDEX IF NOT EXISTS "menus_is_active_idx" ON "menus" ("is_active");
			CREATE INDEX IF NOT EXISTS "menus_is_visible_idx" ON "menus" ("is_visible")
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
