package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gopkg.in/validator.v2"
)

type TrackRepo interface {
	CreateTrack(ctx context.Context, title string, artistId int, filePath string) (*Track, error)
	GetTracks(ctx context.Context, limit int8, offset int8) ([]*Track, error)
	GetTrackByTitle(ctx context.Context, title string) ([]*Track, error)
	GetTrackByArtist(ctx context.Context, artist string) ([]*Track, error)
	BulkCreateTracks(ctx context.Context, trackTitles []string, artistId int, fileNames []string) (int64, error)
}

type trackRepo struct {
	Db DbConn
}

func NewTrackRepo(db DbConn) TrackRepo {
	return &trackRepo{
		Db: db,
	}
}

func (r *trackRepo) CreateTrack(ctx context.Context, title string, artistId int, filePath string) (*Track, error) {
	track := &Track{}
	track.ArtistId = artistId
	track.Title = title
	track.File = filePath

	if err := validator.Validate(track); err != nil {
		return nil, err
	}

	stmt := `INSERT INTO auxstream.tracks (title, artist_id, file)
             VALUES ($1, $2, $3)
             RETURNING id, created_at
             `
	row := r.Db.QueryRow(ctx, stmt, track.Title, track.ArtistId, track.File)

	err := row.Scan(&track.Id, &track.CreatedAt)
	return track, err
}

func (r *trackRepo) GetTracks(ctx context.Context, limit int8, offset int8) ([]*Track, error) {
	tracks := []*Track{}
	stmt := `SELECT id, title, artist_id, file, created_at
			 FROM auxstream.tracks
			 LIMIT $1
			 OFFSET $2
			 `
	rows, err := r.Db.Query(ctx, stmt, limit, offset)
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

func (r *trackRepo) GetTrackByTitle(ctx context.Context, title string) ([]*Track, error) {
	tracks := []*Track{}
	stmt := `SELECT id, title, artist_id, file, created_at
			 FROM auxstream.tracks
			 WHERE title = $1
			 `
	rows, err := r.Db.Query(ctx, stmt, title)
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

func (r *trackRepo) GetTrackByArtist(ctx context.Context, artist string) ([]*Track, error) {
	tracks := []*Track{}
	stmt := `SELECT t.id, t.title, t.artist_id, t.file, t.created_at
	FROM auxstream.tracks AS t
	JOIN auxstream.artists AS a ON t.artist_id = a.id
	WHERE a.name = $1
	`
	rows, err := r.Db.Query(ctx, stmt, artist)
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

func (r *trackRepo) BulkCreateTracks(ctx context.Context, trackTitles []string, artistId int, fileNames []string) (int64, error) {
	var rows [][]interface{}

	for idx, title := range trackTitles {
		rows = append(rows, []interface{}{title, artistId, fileNames[idx]})
	}
	count, err := r.Db.CopyFrom(
		ctx,
		pgx.Identifier{"auxstream", "tracks"},
		[]string{"title", "artist_id", "file"},
		pgx.CopyFromRows(rows),
	)
	return count, err
}
