package migrations

import (
	"context"

	"webapi/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createOrganizationsTable20250115000000)
}

var createOrganizationsTable20250115000000 = &Migration{
	Name: "20250115000000_create_organizations_table",
	Up: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS organizations (
				"id" UUID NOT NULL,
				"name" VARCHAR(255) NOT NULL,
				"description" TEXT,
				"code" VARCHAR(50) UNIQUE,
				"email" VARCHAR(255),
				"phone" VARCHAR(50),
				"address" TEXT,
				"website" VARCHAR(255),
				"logo_url" VARCHAR(500),
				"status" VARCHAR(20) DEFAULT 'active',
				"parent_id" UUID,
				"record_left" INTEGER NULL,
				"record_right" INTEGER NULL,
				"record_depth" INTEGER NULL,
				"record_ordering" INTEGER NULL,
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"deleted_at" TIMESTAMP WITH TIME ZONE,
				CONSTRAINT "organizations_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "organizations_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "organizations"("id") ON DELETE SET NULL,
				CONSTRAINT "organizations_status_check" CHECK ("status" IN ('active', 'inactive', 'suspended')),
				CONSTRAINT "organizations_record_left_right_check" CHECK ("record_left" < "record_right"),
				CONSTRAINT "organizations_record_ordering_check" CHECK ("record_ordering" >= 0)
			);

			-- Create indexes for nested set operations
			CREATE INDEX IF NOT EXISTS "organizations_record_left_idx" ON "organizations" ("record_left");
			CREATE INDEX IF NOT EXISTS "organizations_record_right_idx" ON "organizations" ("record_right");
			CREATE INDEX IF NOT EXISTS "organizations_record_ordering_idx" ON "organizations" ("record_ordering");
			CREATE INDEX IF NOT EXISTS "organizations_parent_id_idx" ON "organizations" ("parent_id");
			CREATE INDEX IF NOT EXISTS "organizations_status_idx" ON "organizations" ("status");
			CREATE INDEX IF NOT EXISTS "organizations_code_idx" ON "organizations" ("code");
			CREATE INDEX IF NOT EXISTS "organizations_deleted_at_idx" ON "organizations" ("deleted_at");
		`)
		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS organizations;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
