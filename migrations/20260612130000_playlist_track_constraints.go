package migrations

import (
	"time"

	"github.com/beesaferoot/gorm-migrate/migration"
	"gorm.io/gorm"
)

func init() {
	migration.RegisterMigration(&migration.Migration{
		Version:   "20260612130000",
		Name:      "playlist_track_constraints",
		CreatedAt: time.Now(),
		// Back the playlist feature: prevent the same track being added to a
		// playlist twice (partial unique so a track can be re-added after a
		// soft-delete), and index playlist ordering for fast track listing.
		Up: func(db *gorm.DB) error {
			if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_auxstream_playlist_tracks_unique
				ON "auxstream"."playlist_tracks" ("playlist_id", "track_id")
				WHERE deleted_at IS NULL;`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_auxstream_playlist_tracks_order
				ON "auxstream"."playlist_tracks" ("playlist_id", "position");`).Error; err != nil {
				return err
			}
			return nil
		},
		Down: func(db *gorm.DB) error {
			if err := db.Exec(`DROP INDEX IF EXISTS "auxstream"."idx_auxstream_playlist_tracks_order";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP INDEX IF EXISTS "auxstream"."idx_auxstream_playlist_tracks_unique";`).Error; err != nil {
				return err
			}
			return nil
		},
	})
}
