package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
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
			    "record_left" BIGINT NULL,
			    "record_right" BIGINT NULL,
			    "record_depth" BIGINT NULL,
			    "record_ordering" BIGINT NULL,
			    "parent_id" UUID NULL,
			    "created_by" UUID NULL,
			    "updated_by" UUID NULL,
			    "deleted_by" UUID NULL,
			    "deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
			    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			    -- Foreign key constraints removed to allow NULL values for user fields
			    -- CONSTRAINT "taxonomies_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    -- CONSTRAINT "taxonomies_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    -- CONSTRAINT "taxonomies_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "taxonomies_parent_id_foreign" FOREIGN KEY ("parent_id") REFERENCES "taxonomies" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "taxonomies_record_left_right_check" CHECK ("record_left" < "record_right"),
				CONSTRAINT "taxonomies_record_ordering_check" CHECK ("record_ordering" >= 0),
				CONSTRAINT "taxonomies_record_depth_check" CHECK ("record_depth" >= 0)
			)
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
