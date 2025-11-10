package migrations

import (
	"github.com/beesaferoot/gorm-migrate/migration"
	"gorm.io/gorm"
	"time"
)

func init() {
	migration.RegisterMigration(&migration.Migration{
		Version:   "20251109110457",
		Name:      "init_db_schema",
		CreatedAt: time.Now(),
		Up: func(db *gorm.DB) error {
			if err := db.Exec(`CREATE SCHEMA IF NOT EXISTS auxstream AUTHORIZATION CURRENT_USER;`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream"."artists" (
	deleted_at timestamp,
	id uuid
	PRIMARY KEY,
	name varchar(255)
	NOT NULL,
	created_at timestamp,
	updated_at timestamp
);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_artists_deleted_at ON "auxstream"."artists" ("deleted_at");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_artists_name ON "auxstream"."artists" ("name");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream"."tracks" (
id uuid
	PRIMARY KEY,
	title varchar(255)
	NOT NULL,
	artist_id uuid
	NOT NULL,
	file varchar(255)
	NOT NULL,
	duration integer
	DEFAULT 0,
	thumbnail text,
	created_at timestamp,
	deleted_at timestamp,
	updated_at timestamp,
	CONSTRAINT "fk_auxstream.tracks_artist_id_fkey"
		FOREIGN KEY ("artist_id")
		REFERENCES "auxstream"."artists"(id)
		ON DELETE CASCADE
	);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_tracks_deleted_at ON "auxstream"."tracks" ("deleted_at");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream"."users" (
id uuid
	PRIMARY KEY,
	email varchar(255),
	password_hash varchar(255),
	google_id varchar(255),
	provider varchar(255) DEFAULT 'local',
	created_at timestamp,
	updated_at timestamp,
	deleted_at timestamp
);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_users_email ON "auxstream"."users" ("email");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_users_google_id ON "auxstream"."users" ("google_id");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_users_deleted_at ON "auxstream"."users" ("deleted_at");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream"."playlists" (
user_id uuid
	NOT NULL,
	name varchar(255)
	NOT NULL,
	description text,
	is_public boolean
	DEFAULT false,
	created_at timestamp,
	deleted_at timestamp,
	id uuid
	PRIMARY KEY,
	updated_at timestamp,
	CONSTRAINT "fk_auxstream.playlists_user_id_fkey"
		FOREIGN KEY ("user_id")
		REFERENCES "auxstream"."users"(id)
		ON DELETE CASCADE
	);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_playlists_deleted_at ON "auxstream"."playlists" ("deleted_at");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream"."playlist_tracks" (
	deleted_at timestamp,
	id uuid
	PRIMARY KEY,
	playlist_id uuid
	NOT NULL,
	track_id uuid
	NOT NULL,
	position integer
	DEFAULT 0,
	added_at timestamp,
	CONSTRAINT "fk_auxstream.playlist_tracks_track_id_fkey"
		FOREIGN KEY ("track_id")
		REFERENCES "auxstream"."tracks"(id)
		ON DELETE CASCADE,
	CONSTRAINT "fk_auxstream.playlist_tracks_playlist_id_fkey"
		FOREIGN KEY ("playlist_id")
		REFERENCES "auxstream"."playlists"(id)
		ON DELETE CASCADE
	);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_playlist_tracks_deleted_at ON "auxstream"."playlist_tracks" ("deleted_at");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream"."playback_history" (
user_id uuid
	NOT NULL,
	track_id uuid
	NOT NULL,
	played_at timestamp,
	duration_played integer,
	deleted_at timestamp,
	id uuid
	PRIMARY KEY,
	CONSTRAINT "fk_auxstream.playback_history_user_id_fkey"
		FOREIGN KEY ("user_id")
		REFERENCES "auxstream"."users"(id)
		ON DELETE CASCADE,
	CONSTRAINT "fk_auxstream.playback_history_track_id_fkey"
		FOREIGN KEY ("track_id")
		REFERENCES "auxstream"."tracks"(id)
		ON DELETE CASCADE
	);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_playback_history_deleted_at ON "auxstream"."playback_history" ("deleted_at");`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream"."track_sources" (
id uuid
	PRIMARY KEY,
	stream_url text,
	created_at timestamp,
	updated_at timestamp,
	deleted_at timestamp,
	track_id uuid
	NOT NULL,
	source varchar(255)
	NOT NULL,
	external_id varchar(255),
	duration integer,
	CONSTRAINT "fk_auxstream.track_sources_track_id_fkey"
		FOREIGN KEY ("track_id")
		REFERENCES "auxstream"."tracks"(id)
		ON DELETE CASCADE
	);`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE INDEX idx_auxstream_track_sources_deleted_at ON "auxstream"."track_sources" ("deleted_at");`).Error; err != nil {
				return err
			}
			return nil
		},
		Down: func(db *gorm.DB) error {
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream".track_sources;`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream"."playback_history";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream"."playlist_tracks";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream"."playlists";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream"."users";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream"."tracks";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream"."artists";`).Error; err != nil {
				return err
			}
			if err := db.Exec(`DROP SCHEMA IF EXISTS auxstream CASCADE;`).Error; err != nil {
				return err
			}
			return nil
		},
	})
}
