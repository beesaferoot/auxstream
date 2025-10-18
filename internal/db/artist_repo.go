package db

import (
	"context"

	"github.com/google/uuid"
	"gopkg.in/validator.v2"
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

func (r *artistRepo) CreateArtist(ctx context.Context, name string) (*Artist, error) {
	artist := &Artist{Name: name}

	if err := validator.Validate(artist); err != nil {
		return nil, err
	}

	// Use GORM's FirstOrCreate to handle the "ON CONFLICT DO NOTHING" logic
	res := r.Db.WithContext(ctx).Where("name = ?", name).FirstOrCreate(artist)

	return artist, res.Error
}

func (r *artistRepo) GetArtistById(ctx context.Context, id uuid.UUID) (*Artist, error) {
	artist := &Artist{}
	res := r.Db.WithContext(ctx).First(artist, id)
	return artist, res.Error
}
