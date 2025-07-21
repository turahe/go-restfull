package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createTagsTable)
}

var createTagsTable = &Migration{
	Name: "20250708231759_create_tags_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS tags (
				"id" UUID NOT NULL,
				"name" varchar(255) NOT NULL,
				"slug" varchar(255) NOT NULL UNIQUE,
				"created_by" UUID NULL,
				"updated_by" UUID NULL,
				"deleted_by" UUID NULL,
				"deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "tags_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "tags_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "tags_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION
			);
			CREATE TABLE IF NOT EXISTS taggables (
				"id" UUID NOT NULL,
				"tag_id" UUID NOT NULL,
				"taggable_id" UUID NOT NULL,
				"taggable_type" varchar(255) NOT NULL,
				"created_at" BIGINT NULL
			);
		`)

		if err != nil {
			return err
		}
		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS tags;
			DROP TABLE IF EXISTS taggables;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
