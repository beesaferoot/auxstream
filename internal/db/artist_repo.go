package db

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArtistRepo interface {
	CreateArtist(ctx context.Context, name string) (*Artist, error)
	GetArtistById(ctx context.Context, id uuid.UUID) (*Artist, error)
}

type artistRepo struct {
	Db *gorm.DB
}

func NewArtistRepo(db *gorm.DB) ArtistRepo {
	return &artistRepo{
		Db: db,
	}
}

// CreateArtist returns the existing artist with this name, or creates one if
// none exists; it never produces a duplicate, so it is safe to call per upload.
func (r *artistRepo) CreateArtist(ctx context.Context, name string) (*Artist, error) {
	artist := &Artist{}

	res := r.Db.WithContext(ctx).Where("name = ?", name).
		Attrs(Artist{ID: uuid.New(), Name: name}).
		FirstOrCreate(artist)

	return artist, res.Error
}

func (r *artistRepo) GetArtistById(ctx context.Context, id uuid.UUID) (*Artist, error) {
	artist := &Artist{}
	res := r.Db.WithContext(ctx).First(artist, id)
	return artist, res.Error
}
