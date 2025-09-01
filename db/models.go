package db

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null" validate:"min=4,nonzero"`
	PasswordHash string         `json:"password_hash" gorm:"not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (User) TableName() string {
	return "auxstream.users"
}

type Track struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"not null" validate:"nonzero"`
	ArtistID  uint           `json:"artist_id" gorm:"not null"`
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
	ID        uint           `json:"id" gorm:"primaryKey"`
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
