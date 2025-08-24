package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createAddressesTable)
}

var createAddressesTable = &Migration{
	Name: "20250729213850_create_addresses_table",
	Up: func() error {
		// Create the addresses table
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS addresses (
				"id" UUID NOT NULL,
				"addressable_id" UUID NOT NULL,
				"addressable_type" VARCHAR(50) NOT NULL,
				"address_line1" VARCHAR(255) NOT NULL,
				"address_line2" VARCHAR(255),
				"city" VARCHAR(255) NOT NULL,
				"state" VARCHAR(255) NOT NULL,
				"province" VARCHAR(255) NULL,
				"regency" VARCHAR(255)  NULL,
				"district" VARCHAR(255) NULL,
				"sub_district" VARCHAR(255) NULL,
				"village" VARCHAR(255) NULL,
				"street" VARCHAR(255) NULL,
				"ward" VARCHAR(255) NULL,
				"postal_code" VARCHAR(20) NOT NULL,
				"country" VARCHAR(255) NOT NULL,
				"latitude" DECIMAL(10, 8),
				"longitude" DECIMAL(11, 8),
				"is_primary" BOOLEAN DEFAULT FALSE,
				"address_type" VARCHAR(50) DEFAULT 'home',
				"created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				"deleted_at" TIMESTAMP WITH TIME ZONE,
				CONSTRAINT "addresses_pkey" PRIMARY KEY ("id"),
				CONSTRAINT "addresses_addressable_type_check" CHECK ("addressable_type" IN ('user', 'organization')),
				CONSTRAINT "addresses_address_type_check" CHECK ("address_type" IN ('home', 'work', 'billing', 'shipping', 'other'))
			)
		`)

		if err != nil {
			return err
		}

		// Create indexes for better performance separately
		indexes := []string{
			`CREATE INDEX IF NOT EXISTS "addresses_addressable_id_idx" ON "addresses" ("addressable_id")`,
			`CREATE INDEX IF NOT EXISTS "addresses_addressable_type_idx" ON "addresses" ("addressable_type")`,
			`CREATE INDEX IF NOT EXISTS "addresses_addressable_composite_idx" ON "addresses" ("addressable_id", "addressable_type")`,
			`CREATE INDEX IF NOT EXISTS "addresses_is_primary_idx" ON "addresses" ("is_primary")`,
			`CREATE INDEX IF NOT EXISTS "addresses_address_type_idx" ON "addresses" ("address_type")`,
			`CREATE INDEX IF NOT EXISTS "addresses_deleted_at_idx" ON "addresses" ("deleted_at")`,
		}

		for _, indexSQL := range indexes {
			_, err := pgx.GetPgxPool().Exec(context.Background(), indexSQL)
			if err != nil {
				return err
			}
		}

		return nil

	},
	Down: func() error {
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			DROP TABLE IF EXISTS addresses;
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
