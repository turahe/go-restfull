package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createMediaTable)
}

var createMediaTable = &Migration{
	Name: "20250507232036_create_media_table",
	Up: func() error {
		// Create the media table
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS media (
			    "id" UUID NOT NULL,
			    "name" varchar(255)  NOT NULL,
			    "hash" varchar(255),
			    "file_name" varchar(255)  NOT NULL,
			    "disk" varchar(255)  NOT NULL,
			    "mime_type" varchar(255)  NOT NULL,
			    "size" int4 NOT NULL,
			    "record_left" BIGINT NULL,
			    "record_right" BIGINT NULL,
			    "record_depth" BIGINT NULL,
			    "record_ordering" BIGINT NULL,
			    "parent_id" UUID NULL,
			    "custom_attributes" varchar(255),
			    "created_by" UUID NULL,
			    "updated_by" UUID NULL,
			    "deleted_by" UUID NULL,
			    "deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
			    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			    CONSTRAINT "media_pkey" PRIMARY KEY ("id"),
			    CONSTRAINT "media_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "media_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "media_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "media_parent_id_foreign" FOREIGN KEY ("parent_id") REFERENCES "media" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "media_record_left_right_check" CHECK ("record_left" < "record_right"),
				CONSTRAINT "media_record_ordering_check" CHECK ("record_ordering" >= 0),
				CONSTRAINT "media_record_depth_check" CHECK ("record_depth" >= 0)
			)
		`)

		if err != nil {
			return err
		}

		// Create indexes for nested set operations
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE INDEX IF NOT EXISTS "media_record_left_idx" ON "media" ("record_left");
			CREATE INDEX IF NOT EXISTS "media_record_right_idx" ON "media" ("record_right");
			CREATE INDEX IF NOT EXISTS "media_record_ordering_idx" ON "media" ("record_ordering");
			CREATE INDEX IF NOT EXISTS "media_record_depth_idx" ON "media" ("record_depth");
			CREATE INDEX IF NOT EXISTS "media_parent_id_idx" ON "media" ("parent_id")
		`)

		if err != nil {
			return err
		}

		// Create the mediables table
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS mediables (
			    "media_id" UUID NOT NULL,
			    "mediable_id" UUID NOT NULL,
			    "mediable_type" varchar(255)  NOT NULL,
			    "group" varchar(255)  NOT NULL
			)
		`)

		if err != nil {
			return err
		}

		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS mediables;
			DROP TABLE IF EXISTS media;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
