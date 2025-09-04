package migrations

import (
	"github.com/beesaferoot/gorm-schema/migration"
	"gorm.io/gorm"
	"time"
)

func init() {
	migration.RegisterMigration(&migration.Migration{
		Version:   "20250904224350",
		Name:      "init_db_schema",
		CreatedAt: time.Now(),
		Up: func(db *gorm.DB) error {
			if err := db.Exec(`CREATE SCHEMA IF NOT EXISTS auxstream;`).Error; err != nil {
				return err
			}
			if err := db.Exec(`CREATE TABLE "auxstream.artists" (
	created_at timestamp,
	updated_at timestamp,
	deleted_at timestamp,
	id BIGSERIAL
	PRIMARY KEY,
	name varchar(255)
	NOT NULL
);`).Error; err != nil {
			return err
		}
		if err := db.Exec(`CREATE INDEX idx_artist_name ON "auxstream.artists" ("name");`).Error; err != nil {
			return err
		}
		if err := db.Exec(`CREATE INDEX idx_artist_deleted_at ON "auxstream.artists" ("deleted_at");`).Error; err != nil {
			return err
		}
		if err := db.Exec(`CREATE TABLE "auxstream.tracks" (
	file varchar(255)
	NOT NULL,
	created_at timestamp,
	updated_at timestamp,
	deleted_at timestamp,
	id BIGSERIAL
	PRIMARY KEY,
	title varchar(255)
	NOT NULL,
	artist_id bigint
	NOT NULL,
	CONSTRAINT fk_track_artist_id_fkey
		FOREIGN KEY ("artist_id")
		REFERENCES "auxstream.artists"(id)
		ON DELETE CASCADE
	);`).Error; err != nil {
			return err
		}
		if err := db.Exec(`CREATE INDEX idx_track_deleted_at ON "auxstream.tracks" ("deleted_at");`).Error; err != nil {
			return err
		}
		if err := db.Exec(`CREATE TABLE "auxstream.users" (
	updated_at timestamp,
	deleted_at timestamp,
	id BIGSERIAL
	PRIMARY KEY,
	username varchar(255) NOT NULL,
	password_hash varchar(255) NOT NULL,
	created_at timestamp
);`).Error; err != nil {
			return err
		}
		if err := db.Exec(`CREATE INDEX idx_user_username ON "auxstream.users" ("username");`).Error; err != nil {
			return err
		}
		if err := db.Exec(`CREATE INDEX idx_user_deleted_at ON "auxstream.users" ("deleted_at");`).Error; err != nil {
			return err
		}
			return nil
		},
		Down: func(db *gorm.DB) error {
			if err := db.Exec(`DROP TABLE IF EXISTS "auxstream.users";`).Error; err != nil {
			return err
		}
		if err := db.Exec(`DROP TABLE IF EXISTS "auxstream.tracks";`).Error; err != nil {
			return err
		}
		if err := db.Exec(`DROP TABLE IF EXISTS "auxstream.artists";`).Error; err != nil {
			return err
		}
		if  err := db.Exec(`DROP SCHEMA IF EXISTS auxstream CASCADE;`).Error; err != nil {
			return err
		}
			return nil
		},
	})
}
