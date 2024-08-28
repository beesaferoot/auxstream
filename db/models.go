package db

import (
	"encoding/json"
	"time"
)

type User struct {
	Id           int       `json:"id"`
	Username     string    `json:"username" validate:"min=4,nonzero"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type Track struct {
	Id        int       `json:"id"`
	Title     string    `json:"title" validate:"nonzero"`
	ArtistId  int       `json:"artist_id"`
	File      string    `json:"file"`
	CreatedAt time.Time `json:"created_at"`
}

type Artist struct {
	Id        int       `json:"id"`
	Name      string    `json:"name" validate:"nonzero"`
	CreatedAt time.Time `json:"created_at"`
}

func (user *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(user)
}

func (user *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, user)
}

func (track *Track) MarshalBinary() ([]byte, error) {
	return json.Marshal(track)
}

func (track *Track) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, track)
}

func (artist *Artist) MarshalBinary() ([]byte, error) {
	return json.Marshal(artist)
}

func (artist *Artist) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, artist)
}


