package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, createActivitiesTable)
}

var createActivitiesTable = &Migration{
	Name: "20250810161536_create_activities_table",
	Up: func() error {
		// Create the table first
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			CREATE TABLE activities (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL,
				action VARCHAR(255) NOT NULL,
				model_type VARCHAR(255) NOT NULL,
				model_id UUID NOT NULL,
				description TEXT,
				properties JSONB,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMP WITH TIME ZONE,
				CONSTRAINT fk_activities_user FOREIGN KEY (user_id) REFERENCES users(id)
			)
		`)

		if err != nil {
			return err
		}

		// Create indexes separately
		indexes := []string{
			`CREATE INDEX IF NOT EXISTS "activities_user_id_idx" ON "activities" ("user_id")`,
			`CREATE INDEX IF NOT EXISTS "activities_model_type_idx" ON "activities" ("model_type")`,
			`CREATE INDEX IF NOT EXISTS "activities_model_id_idx" ON "activities" ("model_id")`,
			`CREATE INDEX IF NOT EXISTS "activities_deleted_at_idx" ON "activities" ("deleted_at")`,
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
			DROP TABLE activities
		`)
		if err != nil {
			return err
		}

		return nil
	},
}
