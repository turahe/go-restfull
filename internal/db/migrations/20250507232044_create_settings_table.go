package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createSettingTable)
}

var createSettingTable = &Migration{
	Name: "20250507232044_create_settings_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE settings (
				 "id" UUID NOT NULL,
				"model_type" VARCHAR(255) NULL,
				"model_id" UUID NULL,
			    "key" VARCHAR(255) NOT NULL,
			    "value" VARCHAR(255) NULL,
			    "created_by" UUID NULL,
			    "updated_by" UUID NULL,
			    "deleted_by" UUID NULL,
			    "deleted_at" BIGINT NULL,
			    "created_at" BIGINT NULL,
			    "updated_at" BIGINT NULL,
			    CONSTRAINT "setting_pkey" PRIMARY KEY ("id"),
			    CONSTRAINT "setting_created_by_foreign" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "setting_deleted_by_foreign" FOREIGN KEY ("deleted_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
			    CONSTRAINT "setting_updated_by_foreign" FOREIGN KEY ("updated_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION
			    
			);
		`)

		if err != nil {
			return err
		}
		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS settings;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
