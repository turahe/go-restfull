package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createTaxonomyTable)
}

var createTaxonomyTable = &Migration{
	Name: "20250510195724_create_taxonomies_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS taxonomies (
			    "id" UUID NOT NULL PRIMARY KEY,
			    "name" varchar(255) NOT NULL,
			    "slug" varchar(255) NOT NULL UNIQUE,
			    "code" varchar(255),
			    "description" text,
			    "record_left" int8,
			    "record_right" int8,
			    "record_ordering" int8,
			    "parent_id" UUID NULL,
			    "created_by" UUID NULL,
			    "updated_by" UUID NULL,
			    "deleted_by" UUID NULL,
			    "deleted_at" BIGINT NULL,
			    "created_at" BIGINT NULL,
			    "updated_at" BIGINT NULL,
			    CONSTRAINT "taxonomies_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "taxonomies_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "taxonomies_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION
			);
		`)

		if err != nil {
			return err
		}
		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS taxonomies;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
