package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createContentTable)
}

var createContentTable = &Migration{
	Name: "20250510193942_create_contents_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS contents (
			    "id" UUID NOT NULL,
			    "model_type" varchar(255) NOT NULL,
			    "model_id" UUID NOT NULL,
			    "content_raw" text NOT NULL,
			    "content_html" text NOT NULL,
			    "created_by" UUID NULL,
			    "updated_by" UUID NULL,
			    "deleted_by" UUID NULL,
			    "deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
			    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			    CONSTRAINT "contents_pkey" PRIMARY KEY ("id"),
			    CONSTRAINT "contents_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "contents_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "contents_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION
			);
		`)

		if err != nil {
			return err
		}
		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS contents;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
