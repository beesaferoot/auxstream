package db

import (
	"context"

	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type TrackRepo interface {
	CreateTrack(ctx context.Context, title string, artistId int, filePath string) (*Track, error)
	GetTracks(ctx context.Context, limit int, offset int) ([]*Track, error)
	GetTrackByTitle(ctx context.Context, title string) ([]*Track, error)
	GetTrackByArtist(ctx context.Context, artist string) ([]*Track, error)
	BulkCreateTracks(ctx context.Context, trackTitles []string, artistId uint, fileNames []string) (int64, error)
}

type trackRepo struct {
	Db *gorm.DB
}

func NewTrackRepo(db *gorm.DB) TrackRepo {
	return &trackRepo{
		Db: db,
	}
}

func (r *trackRepo) CreateTrack(ctx context.Context, title string, artistId int, filePath string) (*Track, error) {
	track := &Track{
		Title:    title,
		ArtistID: uint(artistId),
		File:     filePath,
	}

	if err := validator.Validate(track); err != nil {
		return nil, err
	}

	res := r.Db.WithContext(ctx).Create(track)

	return track, res.Error
}

func (r *trackRepo) GetTracks(ctx context.Context, limit int, offset int) ([]*Track, error) {
	var tracks []*Track

	res := r.Db.WithContext(ctx).Limit(limit).Offset(offset).Find(&tracks)

	if res.Error != nil {
		return tracks, res.Error
	}

	return tracks, nil
}

func (r *trackRepo) GetTrackByTitle(ctx context.Context, title string) ([]*Track, error) {
	var tracks []*Track
	res := r.Db.WithContext(ctx).Preload("Artist").Where("title ILIKE ?", "%"+title+"%").Find(&tracks)

	if res.Error != nil {
		return tracks, res.Error
	}

	return tracks, nil
}

func (r *trackRepo) GetTrackByArtist(ctx context.Context, artist string) ([]*Track, error) {
	var tracks []*Track
	res := r.Db.WithContext(ctx).Joins("JOIN auxstream.artists ON auxstream.tracks.artist_id = auxstream.artists.id").
		Where("auxstream.artists.name ILIKE ?", "%"+artist+"%").
		Find(&tracks)

	if res.Error != nil {
		return tracks, res.Error
	}

	return tracks, nil
}

func (r *trackRepo) BulkCreateTracks(ctx context.Context, trackTitles []string, artistId uint, fileNames []string) (int64, error) {
	var tracks []Track

	for idx, title := range trackTitles {
		tracks = append(tracks, Track{
			Title:    title,
			ArtistID: artistId,
			File:     fileNames[idx],
		})
	}

	res := r.Db.WithContext(ctx).CreateInBatches(tracks, 100)

	return res.RowsAffected, res.Error
}
