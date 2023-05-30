package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

type Track struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	ArtistId  int       `json:"artist_id"`
	File      string    `json:"file"`
	CreatedAt time.Time `json:"created_at"`
}

type Artist struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (track *Track) Commit(ctx context.Context, trx pgx.Tx) (err error) {
	stmt := `INSERT INTO auxstream.tracks (title, artist_id, file) 
             VALUES ($1, $2, $3) 
             RETURNING id, created_at
             `
	row := trx.QueryRow(ctx, stmt, track.Title, track.ArtistId, track.File)

	err = row.Scan(&track.Id, &track.CreatedAt)
	return
}

func (artist *Artist) Commit(ctx context.Context, trx pgx.Tx) (err error) {
	stmt := `INSERT INTO auxstream.artists (name) 
			 VALUES ($1) 
			 RETURNING id, created_at
			 `
	row := trx.QueryRow(ctx, stmt, artist.Name)

	err = row.Scan(&artist.Id, &artist.CreatedAt)
	return
}

func GetTrackByTitle(ctx context.Context, title string) (tracks []*Track, err error) {
	stmt := `SELECT id, title, artist_id, file, created_at 
			 FROM auxstream.tracks 
			 WHERE title = $1
			 `
	rows, err := DAO.conn.Query(ctx, stmt, title)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		track := &Track{}
		err = rows.Scan(&track.Id, &track.Title, &track.ArtistId, &track.CreatedAt)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

func GetTrackByArtist(ctx context.Context, artist string) (tracks []*Track, err error) {
	stmt := `SELECT t.id, t.title, t.artist_id, t.file, t.created_at
	FROM auxstream.tracks AS t
	JOIN auxstream.artists AS a ON t.artist_id = a.id
	WHERE a.name = $1
	`
	rows, err := DAO.conn.Query(ctx, stmt, artist)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		track := &Track{}
		err = rows.Scan(&track.Id, &track.Title, &track.ArtistId, &track.File, &track.CreatedAt)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return tracks, nil
}
