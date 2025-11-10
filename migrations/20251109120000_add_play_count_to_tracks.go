package migrations

import (
	"time"

	"github.com/beesaferoot/gorm-migrate/migration"
	"gorm.io/gorm"
)

func init() {
	migration.RegisterMigration(&migration.Migration{
		Version:   "20251109120000",
		Name:      "add_play_count_to_tracks",
		CreatedAt: time.Now(),
		Up: func(db *gorm.DB) error {
			if err := db.Exec(`ALTER TABLE "auxstream"."tracks" 
				ADD COLUMN IF NOT EXISTS play_count integer DEFAULT 0;`).Error; err != nil {
				return err
			}

			if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_auxstream_tracks_play_count 
				ON "auxstream"."tracks" ("play_count" DESC);`).Error; err != nil {
				return err
			}

			if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_auxstream_tracks_trending 
				ON "auxstream"."tracks" ("play_count" DESC, "created_at" DESC);`).Error; err != nil {
				return err
			}

			return nil
		},
		Down: func(db *gorm.DB) error {
			// Drop indexes
			if err := db.Exec(`DROP INDEX IF EXISTS "auxstream"."idx_auxstream_tracks_trending";`).Error; err != nil {
				return err
			}

			if err := db.Exec(`DROP INDEX IF EXISTS "auxstream"."idx_auxstream_tracks_play_count";`).Error; err != nil {
				return err
			}

			// Remove play_count column
			if err := db.Exec(`ALTER TABLE "auxstream"."tracks" 
				DROP COLUMN IF EXISTS play_count;`).Error; err != nil {
				return err
			}

			return nil
		},
	})
}
