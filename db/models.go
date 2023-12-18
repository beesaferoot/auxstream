package db

import (
	"context"
	"github.com/jackc/pgx/v5"
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

func (user *User) Commit(ctx context.Context) (err error) {
	stmt := `INSERT INTO auxstream.users (username, password_hash)
			 VALUES ($1, $2) 
			 RETURNING id, created_at
			 `
	row := DAO.conn.QueryRow(ctx, stmt, user.Username, user.PasswordHash)
	err = row.Scan(&user.Id, &user.CreatedAt)
	return
}

func (track *Track) Commit(ctx context.Context) (err error) {
	stmt := `INSERT INTO auxstream.tracks (title, artist_id, file) 
             VALUES ($1, $2, $3) 
             RETURNING id, created_at
             `
	row := DAO.conn.QueryRow(ctx, stmt, track.Title, track.ArtistId, track.File)

	err = row.Scan(&track.Id, &track.CreatedAt)
	return
}

func (artist *Artist) Commit(ctx context.Context) (err error) {
	// This query ensures we don't create a new artist record if we already have one
	stmt := `
    INSERT INTO auxstream.artists (name) VALUES ($1)
    ON CONFLICT (name) DO NOTHING
	RETURNING id, created_at
	`
	row := DAO.conn.QueryRow(ctx, stmt, artist.Name)

	err = row.Scan(&artist.Id, &artist.CreatedAt)
	return
}

func GetTracks(ctx context.Context, limit int32, offset int32) (tracks []*Track, err error) {
	tracks = []*Track{}
	stmt := `SELECT id, title, artist_id, file, created_at 
			 FROM auxstream.tracks 
			 LIMIT $1 
			 OFFSET $2
			 `
	rows, err := DAO.conn.Query(ctx, stmt, limit, offset)
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

func GetTrackByTitle(ctx context.Context, title string) (tracks []*Track, err error) {
	tracks = []*Track{}
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

func GetTrackByArtist(ctx context.Context, artist string) (tracks []*Track, err error) {
	tracks = []*Track{}
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

func BulkCreateTrack(ctx context.Context, trackTitles []string, artistId int, fileNames []string) (count int64, err error) {
	var rows [][]interface{}

	for idx, title := range trackTitles {
		rows = append(rows, []interface{}{title, artistId, fileNames[idx]})
	}
	count, err = DAO.conn.CopyFrom(
		ctx,
		pgx.Identifier{"auxstream", "tracks"},
		[]string{"title", "artist_id", "file"},
		pgx.CopyFromRows(rows),
	)
	return count, err
}

func GetUserById(ctx context.Context, id string) (user *User, err error) {
	stmt := `SELECT id, username, password_hash, created_at
 			 FROM auxstream.users
 			 WHERE id = $1`
	row := DAO.conn.QueryRow(ctx, stmt, id)

	err = row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.CreatedAt)

	return
}

func GetUserByUser(ctx context.Context, username string) (user *User, err error) {
	user = &User{}
	stmt := `SELECT id, username, password_hash, created_at
 			 FROM auxstream.users
 			 WHERE username = $1`
	row := DAO.conn.QueryRow(ctx, stmt, username)

	err = row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.CreatedAt)

	return
}
