package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createCommentTable)
}

var createCommentTable = &Migration{
	Name: "20250708232036_create_comment_table",
	Up: func() error {
		// Create the comments table
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS comments (
				"id" UUID NOT NULL,
				"model_type" varchar(255) NOT NULL,
				"model_id" UUID NOT NULL,
				"title" varchar(255) NOT NULL,
				"status" varchar(255) NOT NULL DEFAULT 'pending',
				"parent_id" UUID NULL,
				"record_left" INTEGER NULL,
				"record_right" INTEGER NULL,
				"record_depth" INTEGER NULL,
				"record_ordering" INTEGER NULL,
				"created_by" UUID NULL,
				"updated_by" UUID NULL,
				"deleted_by" UUID NULL,
				"deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				CONSTRAINT "comments_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "comments_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "comments_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "comments_parent_id_foreign" FOREIGN KEY ("parent_id") REFERENCES "comments" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
				CONSTRAINT "comments_record_left_right_check" CHECK ("record_left" < "record_right"),
				CONSTRAINT "comments_record_ordering_check" CHECK ("record_ordering" >= 0),
				CONSTRAINT "comments_record_depth_check" CHECK ("record_depth" >= 0),
				CONSTRAINT "comments_status_check" CHECK ("status" IN ('pending', 'approved', 'rejected', 'spam', 'trash')),
				CONSTRAINT "comments_model_type_check" CHECK ("model_type" IN ('post', 'page', 'comment', 'media', 'taxonomy', 'organization', 'user', 'menu',  'role', 'setting', 'content', 'tag')),
				CONSTRAINT "comments_model_id_check" CHECK ("model_id" IS NOT NULL)
			)
		`)

		if err != nil {
			return err
		}

		// Create indexes for nested set operations
		_, err = pgx.GetPgxPool().Exec(context.Background(), `
			CREATE INDEX IF NOT EXISTS "comments_record_left_idx" ON "comments" ("record_left");
			CREATE INDEX IF NOT EXISTS "comments_record_right_idx" ON "comments" ("record_right");
			CREATE INDEX IF NOT EXISTS "comments_record_ordering_idx" ON "comments" ("record_ordering");
			CREATE INDEX IF NOT EXISTS "comments_record_depth_idx" ON "comments" ("record_depth");
			CREATE INDEX IF NOT EXISTS "comments_parent_id_idx" ON "comments" ("parent_id");
			CREATE INDEX IF NOT EXISTS "comments_status_idx" ON "comments" ("status")
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
