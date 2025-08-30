package migrations

import (
	"context"

	"github.com/turahe/go-restfull/internal/db/pgx"
)

func init() {
	Migrations = append(Migrations, fixMenusConstraints)
}

var fixMenusConstraints = &Migration{
	Name: "20250708232043_fix_menus_constraints",
	Up: func() error {
		// Drop the problematic deleted_by constraint
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			ALTER TABLE menus DROP CONSTRAINT IF EXISTS menus_deleted_by_check;
		`)
		if err != nil {
			return err
		}
		return nil
	},
	Down: func() error {
		// Re-add the constraint if needed to rollback
		_, err := pgx.GetPgxPool().Exec(context.Background(), `
			ALTER TABLE menus ADD CONSTRAINT menus_deleted_by_check CHECK ("deleted_by" IS NOT NULL);
		`)
		if err != nil {
			return err
		}
		return nil
	},
}
