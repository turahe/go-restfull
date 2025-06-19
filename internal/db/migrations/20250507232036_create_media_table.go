package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createMediaTable)
}

var createMediaTable = &Migration{
	Name: "20250507232036_create_media_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS media (
			    "id" UUID NOT NULL,
			    "name" varchar(255)  NOT NULL,
			    "hash" varchar(255),
			    "file_name" varchar(255)  NOT NULL,
			    "disk" varchar(255)  NOT NULL,
			    "mime_type" varchar(255)  NOT NULL,
			    "size" int4 NOT NULL,
			    "record_left" int8,
			    "record_right" int8,
			    "record_dept" int8,
			    "record_ordering" int8,
			    "parent_id" UUID NULL,
			    "custom_attribute" varchar(255),
			    "created_by" UUID NULL,
			    "updated_by" UUID NULL,
			    "deleted_by" UUID NULL,
			    "deleted_at" BIGINT NULL,
			    "created_at" BIGINT NULL,
			    "updated_at" BIGINT NULL,
			    CONSTRAINT "media_pkey" PRIMARY KEY ("id"),
			    CONSTRAINT "media_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "media_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "media_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION
			);
			
			CREATE TABLE IF NOT EXISTS mediables (
			    "media_id" UUID NOT NULL,
			    "mediable_id" UUID NOT NULL,
			    "mediable_type" varchar(255)  NOT NULL,
			    "group" varchar(255)  NOT NULL
			);
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
