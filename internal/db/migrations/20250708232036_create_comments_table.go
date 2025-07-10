package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createCommentTable)
}

var createCommentTable = &Migration{
	Name: "20250708232036_create_comment_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS comments (
				"id" UUID NOT NULL,
				"model_type" varchar(255) NOT NULL,
				"model_id" UUID NOT NULL,
				"title" varchar(255) NOT NULL,
				"status" varchar(255) NOT NULL DEFAULT 'pending',
				"parent_id" UUID NULL,
				"record_left" int8,
				"record_right" int8,
				"record_ordering" int8,
				"created_by" UUID NULL,
				"updated_by" UUID NULL,
				"deleted_by" UUID NULL,
				"deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "comments_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "comments_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "comments_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			)
		`)

		if err != nil {
			return err
		}
		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS comments;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
