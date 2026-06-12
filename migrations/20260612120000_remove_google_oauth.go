package migrations

import (
	"time"

	"github.com/beesaferoot/gorm-migrate/migration"
	"gorm.io/gorm"
)

func init() {
	migration.RegisterMigration(&migration.Migration{
		Version:   "20260612120000",
		Name:      "remove_google_oauth",
		CreatedAt: time.Now(),
		// Google OAuth was removed from the app; drop the now-unused user columns
		// and index that backed it.
		Up: func(db *gorm.DB) error {
			if err := db.Exec(`DROP INDEX IF EXISTS "auxstream"."idx_auxstream_users_google_id";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`ALTER TABLE "auxstream"."users"
				DROP COLUMN IF EXISTS google_id;`).Error; err != nil {
				return err
			}
			if err := db.Exec(`ALTER TABLE "auxstream"."users"
				DROP COLUMN IF EXISTS provider;`).Error; err != nil {
				return err
			}
			return nil
		},
		Down: func(db *gorm.DB) error {
			if err := db.Exec(`ALTER TABLE "auxstream"."users"
				ADD COLUMN IF NOT EXISTS google_id varchar(255);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`ALTER TABLE "auxstream"."users"
				ADD COLUMN IF NOT EXISTS provider varchar(255) DEFAULT 'local';`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_auxstream_users_google_id
				ON "auxstream"."users" ("google_id");`).Error; err != nil {
				return err
			}
			return nil
		},
	})
}
