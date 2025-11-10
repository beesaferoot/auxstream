package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string         `json:"email" gorm:"uniqueIndex" validate:"required,email"`
	PasswordHash string         `json:"password_hash" gorm:""`
	GoogleID     string         `json:"google_id" gorm:"uniqueIndex"`
	Provider     string         `json:"provider" gorm:"default:'local'"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (User) TableName() string {
	return "auxstream.users"
}

type Track struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title     string         `json:"title" gorm:"not null" validate:"required"`
	ArtistID  uuid.UUID      `json:"artist_id" gorm:"type:uuid;not null"`
	Artist    Artist         `json:"artist" gorm:"foreignKey:ArtistID" validate:"-"`
	File      string         `json:"file" gorm:"not null"`
	Duration  int            `json:"duration" gorm:"default:0"`
	Thumbnail string         `json:"thumbnail" gorm:"type:text"`
	PlayCount int            `json:"play_count" gorm:"default:0;index"` // Track popularity
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Track) TableName() string {
	return "auxstream.tracks"
}

// TrackSource represents external or local track sources
type TrackSource struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TrackID    uuid.UUID      `json:"track_id" gorm:"type:uuid;not null"`
	Track      Track          `json:"track" gorm:"foreignKey:TrackID"`
	Source     string         `json:"source" gorm:"not null"` // 'youtube', 'soundcloud', 'local'
	ExternalID string         `json:"external_id"`            // YouTube video ID, SoundCloud ID, etc.
	StreamURL  string         `json:"stream_url" gorm:"type:text"`
	Duration   int            `json:"duration"` // Duration in seconds
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (TrackSource) TableName() string {
	return "auxstream.track_sources"
}

// Playlist represents a user's playlist
type Playlist struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
	User        User            `json:"user" gorm:"foreignKey:UserID"`
	Name        string          `json:"name" gorm:"not null" validate:"required"`
	Description string          `json:"description" gorm:"type:text"`
	IsPublic    bool            `json:"is_public" gorm:"default:false"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
	Tracks      []PlaylistTrack `json:"tracks" gorm:"foreignKey:PlaylistID"`
}

func (Playlist) TableName() string {
	return "auxstream.playlists"
}

// PlaylistTrack represents the many-to-many relationship between playlists and tracks
type PlaylistTrack struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlaylistID uuid.UUID      `json:"playlist_id" gorm:"type:uuid;not null"`
	Playlist   Playlist       `json:"playlist" gorm:"foreignKey:PlaylistID"`
	TrackID    uuid.UUID      `json:"track_id" gorm:"type:uuid;not null"`
	Track      Track          `json:"track" gorm:"foreignKey:TrackID"`
	Position   int            `json:"position" gorm:"default:0"` // Order in playlist
	AddedAt    time.Time      `json:"added_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (PlaylistTrack) TableName() string {
	return "auxstream.playlist_tracks"
}

// PlaybackHistory tracks user listening history
type PlaybackHistory struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User           User           `json:"user" gorm:"foreignKey:UserID"`
	TrackID        uuid.UUID      `json:"track_id" gorm:"type:uuid;not null"`
	Track          Track          `json:"track" gorm:"foreignKey:TrackID"`
	PlayedAt       time.Time      `json:"played_at"`
	DurationPlayed int            `json:"duration_played"` // Duration in seconds
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (PlaybackHistory) TableName() string {
	return "auxstream.playback_history"
}

type Artist struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Tracks    []Track        `json:"tracks" gorm:"foreignKey:ArtistID"`
}

func (Artist) TableName() string {
	return "auxstream.artists"
}

var ModelTypeRegistry = map[string]any{
	"User":            User{},
	"Track":           Track{},
	"Artist":          Artist{},
	"TrackSource":     TrackSource{},
	"Playlist":        Playlist{},
	"PlaylistTrack":   PlaylistTrack{},
	"PlaybackHistory": PlaybackHistory{},
}
