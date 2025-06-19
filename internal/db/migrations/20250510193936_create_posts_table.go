package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createPostTable)
}

var createPostTable = &Migration{
	Name: "20250510193936_create_posts_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS posts (
			    "id" UUID NOT NULL PRIMARY KEY,
			    "slug" varchar(255) NOT NULL UNIQUE,
			    "title" varchar(255) NOT NULL,
			    "subtitle" varchar(255),
			    "description" text,
			    "type" varchar(255) NOT NULL,
			    "is_sticky" bool NOT NULL DEFAULT false,
			    "published_at" BIGINT NULL,
			    "language" varchar(255) NOT NULL DEFAULT 'en'::character varying,
			    "layout" varchar(255),
			    "record_ordering" int8,
			    "created_by" UUID NULL,
			    "updated_by" UUID NULL,
			    "deleted_by" UUID NULL,
			    "deleted_at" BIGINT NULL,
			    "created_at" BIGINT NULL,
			    "updated_at" BIGINT NULL,
			    CONSTRAINT "posts_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "posts_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "posts_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION
			);
		`)

		if err != nil {
			return err
		}
		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS posts;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
