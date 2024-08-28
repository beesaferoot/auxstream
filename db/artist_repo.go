package db

import (
	"context"

	"gopkg.in/validator.v2"
)

type ArtistRepo interface {
	CreateArtist(ctx context.Context, name string) (*Artist, error)
	GetArtistById(ctx context.Context, id int) (*Artist, error)
}

type artistRepo struct {
	Db DbConn
}

func NewArtistRepo(db DbConn) ArtistRepo {
	return &artistRepo{
		Db: db,
	}
}

func (r *artistRepo) CreateArtist(ctx context.Context, name string) (*Artist, error) {
	// This query ensures we don't create a new artist record if we already have one
	artist := &Artist{Name: name}

	if err := validator.Validate(artist); err != nil {
		return nil, err
	}

	stmt := `
    INSERT INTO auxstream.artists (name) VALUES ($1)
    ON CONFLICT (name) DO NOTHING
	RETURNING id, created_at
	`
	row := r.Db.QueryRow(ctx, stmt, artist.Name)

	err := row.Scan(&artist.Id, &artist.CreatedAt)
	return artist, err
}

func (r *artistRepo) GetArtistById(ctx context.Context, id int) (*Artist, error) {
	artist := &Artist{Id: id}
	stmt := `SELECT name, created_at
			 FROM auxstream.artists
			 WHERE id = $1`
	row := r.Db.QueryRow(ctx, stmt, id)

	err := row.Scan(&artist.Name, artist.CreatedAt)

	return artist, err
}
