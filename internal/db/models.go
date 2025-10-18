package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"uniqueIndex" validate:"min=4,nonzero"`
	Email        string         `json:"email" gorm:"uniqueIndex" validate:"email"`
	PasswordHash string         `json:"password_hash" gorm:""`
	GoogleID     string         `json:"google_id" gorm:"uniqueIndex"`
	Provider     string  		`json:"provider" gorm:"type:enum('local', 'google');default:'local'"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (User) TableName() string {
	return "auxstream.users"
}

type Track struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title     string         `json:"title" gorm:"not null" validate:"nonzero"`
	ArtistID  uuid.UUID      `json:"artist_id" gorm:"type:uuid;not null"`
	Artist    Artist         `json:"artist" gorm:"foreignKey:ArtistID" validate:"-"`
	File      string         `json:"file" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Track) TableName() string {
	return "auxstream.tracks"
}

type Artist struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null" validate:"nonzero"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Tracks    []Track        `json:"tracks" gorm:"foreignKey:ArtistID"`
}

func (Artist) TableName() string {
	return "auxstream.artists"
}

var ModelTypeRegistry = map[string]any{
	"User":   User{},
	"Track":  Track{},
	"Artist": Artist{},
}
