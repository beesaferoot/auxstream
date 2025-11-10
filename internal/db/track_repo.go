package db

import (
	"auxstream/internal/logger"
	"context"
	"go.uber.org/zap"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackRepo interface {
	CreateTrack(ctx context.Context, title string, artistId uuid.UUID, filePath string, duration int, thumbnail string) (*Track, error)
	GetTracks(ctx context.Context, limit int, offset int) ([]*Track, error)
	GetTrendingTracks(ctx context.Context, limit int, offset int, days int) ([]*Track, error)
	GetRecentTracks(ctx context.Context, limit int, offset int) ([]*Track, error)
	GetTrackByID(ctx context.Context, id uuid.UUID) (*Track, error)
	GetTrackByTitle(ctx context.Context, title string) ([]*Track, error)
	GetTrackByArtist(ctx context.Context, artist string) ([]*Track, error)
	GetTracksByArtistId(ctx context.Context, artistId uuid.UUID, limit int, offset int) ([]*Track, error)
	SearchTracks(ctx context.Context, query string) ([]*Track, error)
	BulkCreateTracks(ctx context.Context, trackTitles []string, artistId uuid.UUID, fileMetas map[string]string) (int64, error)
	IncrementPlayCount(ctx context.Context, trackId uuid.UUID) error
	RecordPlayback(ctx context.Context, userId uuid.UUID, trackId uuid.UUID, durationPlayed int) error
}

type trackRepo struct {
	Db *gorm.DB
}

func NewTrackRepo(db *gorm.DB) TrackRepo {
	return &trackRepo{
		Db: db,
	}
}

func (r *trackRepo) CreateTrack(ctx context.Context, title string, artistId uuid.UUID, filePath string, duration int, thumbnail string) (*Track, error) {
	track := &Track{
		ID:        uuid.New(),
		Title:     title,
		ArtistID:  artistId,
		File:      filePath,
		Duration:  duration,
		Thumbnail: thumbnail,
	}

	if err := validate.Struct(track); err != nil {
		return nil, err
	}

	res := r.Db.WithContext(ctx).Create(track)

	return track, res.Error
}

func (r *trackRepo) GetTracks(ctx context.Context, limit int, offset int) ([]*Track, error) {
	var tracks []*Track

	res := r.Db.WithContext(ctx).
		Preload("Artist", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "created_at", "updated_at") // only select needed fields
		}).
		Limit(limit).
		Offset(offset).
		Find(&tracks)

	if res.Error != nil {
		return nil, res.Error
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

func (r *trackRepo) GetTracksByArtistId(ctx context.Context, artistId uuid.UUID, limit int, offset int) ([]*Track, error) {
	var tracks []*Track
	res := r.Db.WithContext(ctx).
		Preload("Artist").
		Where("artist_id = ?", artistId).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tracks)

	if res.Error != nil {
		return tracks, res.Error
	}

	return tracks, nil
}

func (r *trackRepo) GetTrackByID(ctx context.Context, id uuid.UUID) (*Track, error) {
	var track Track
	res := r.Db.WithContext(ctx).Preload("Artist").First(&track, "id = ?", id)

	if res.Error != nil {
		return nil, res.Error
	}

	return &track, nil
}

func (r *trackRepo) SearchTracks(ctx context.Context, query string) ([]*Track, error) {
	var tracks []*Track

	res := r.Db.WithContext(ctx).
		Preload("Artist").
		Where("LOWER(title) LIKE LOWER(?)", "%"+query+"%").
		Or("EXISTS (SELECT 1 FROM auxstream.artists WHERE id = tracks.artist_id AND LOWER(name) LIKE LOWER(?))", "%"+query+"%").
		Limit(20).
		Find(&tracks)

	if res.Error != nil {
		return nil, res.Error
	}

	return tracks, nil
}

func (r *trackRepo) BulkCreateTracks(ctx context.Context, trackTitles []string, artistId uuid.UUID, fileNames map[string]string) (int64, error) {
	var tracks []Track

	logger.Info("bulk create tracks", []zap.Field{
		zap.String("artist_id", artistId.String()),
		zap.Strings("track_titles", trackTitles),
		zap.Any("file_names", fileNames),
	}...)

	for _, title := range trackTitles {
		fileName, _ := fileNames[title]
		tracks = append(tracks, Track{
			Title:    title,
			ArtistID: artistId,
			ID:       uuid.New(),
			File:     fileName,
		})
	}

	res := r.Db.WithContext(ctx).CreateInBatches(tracks, 100)

	return res.RowsAffected, res.Error
}

// GetTrendingTracks gets tracks sorted by play count (trending algorithm)
func (r *trackRepo) GetTrendingTracks(ctx context.Context, limit int, offset int, days int) ([]*Track, error) {
	var tracks []*Track

	query := r.Db.WithContext(ctx).
		Preload("Artist", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "created_at", "updated_at")
		})

	// Filter by recency if days is specified
	if days > 0 {
		// Calculate the cutoff date
		cutoffDate := time.Now().AddDate(0, 0, -days)
		query = query.Where("created_at >= ?", cutoffDate)
	}

	res := query.
		Order("play_count DESC").
		Order("created_at DESC"). // Secondary sort by newest
		Limit(limit).
		Offset(offset).
		Find(&tracks)

	if res.Error != nil {
		return nil, res.Error
	}

	return tracks, nil
}

// GetRecentTracks gets recently added tracks
func (r *trackRepo) GetRecentTracks(ctx context.Context, limit int, offset int) ([]*Track, error) {
	var tracks []*Track

	res := r.Db.WithContext(ctx).
		Preload("Artist", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "created_at", "updated_at")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tracks)

	if res.Error != nil {
		return nil, res.Error
	}

	return tracks, nil
}

// IncrementPlayCount increments the play count for a track
func (r *trackRepo) IncrementPlayCount(ctx context.Context, trackId uuid.UUID) error {
	res := r.Db.WithContext(ctx).
		Model(&Track{}).
		Where("id = ?", trackId).
		UpdateColumn("play_count", gorm.Expr("play_count + ?", 1))

	return res.Error
}

// RecordPlayback records a playback event in the history
func (r *trackRepo) RecordPlayback(ctx context.Context, userId uuid.UUID, trackId uuid.UUID, durationPlayed int) error {
	playback := &PlaybackHistory{
		ID:             uuid.New(),
		UserID:         userId,
		TrackID:        trackId,
		PlayedAt:       time.Now(),
		DurationPlayed: durationPlayed,
	}

	res := r.Db.WithContext(ctx).Create(playback)
	return res.Error
}
