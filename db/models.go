package db

import (
	"context"
	"time"
)

type Track struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	ArtistId  int       `json:"artist_id"`
	FileName  string    `json:"file_name"`
	CreatedAt time.Time `json:"created_at"`
}

type Artist struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (track *Track) Commit(ctx context.Context) (err error) {
	stmt := "INSERT INTO auxstream.tracks (title, artist_id, file_name) VALUES ($1, $2, $3) return id, created_at"
	rows := DAO.conn.QueryRow(ctx, stmt, track.Title, track.ArtistId, track.FileName)

	err = rows.Scan(&track.Id, &track.CreatedAt)

	return
}

func (artist *Artist) Commit(ctx context.Context) (err error) {
	stmt := "INSERT INTO auxstream.artists (name) VALUES ($1) return id, created_at"
	rows := DAO.conn.QueryRow(ctx, stmt)

	err = rows.Scan(&artist.Id)

	return
}
