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
	BulkCreateTracks(ctx context.Context, inputs []BulkTrackInput, artistId uuid.UUID) (int64, error)
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
			return db.Select("id", "name", "created_at", "updated_at")
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

// SearchTracks performs a combined title/artist fuzzy search. It is the
// intended single entry point for local search (see .todo: search feature is
// still being developed) and currently complements the aggregator's per-field
// lookups.
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

// BulkTrackInput is a single title/stored-file pair for a bulk upload. Using an
// ordered slice (rather than a title-keyed map) preserves every track even when
// titles repeat.
type BulkTrackInput struct {
	Title string `json:"title"`
	File  string `json:"file"`
}

func (r *trackRepo) BulkCreateTracks(ctx context.Context, inputs []BulkTrackInput, artistId uuid.UUID) (int64, error) {
	if len(inputs) == 0 {
		return 0, nil
	}

	tracks := make([]Track, 0, len(inputs))
	for _, in := range inputs {
		tracks = append(tracks, Track{
			ID:       uuid.New(),
			Title:    in.Title,
			ArtistID: artistId,
			File:     in.File,
		})
	}

	logger.Info("bulk create tracks",
		zap.String("artist_id", artistId.String()),
		zap.Int("count", len(tracks)),
	)

	res := r.Db.WithContext(ctx).CreateInBatches(tracks, 100)

	return res.RowsAffected, res.Error
}

// GetTrendingTracks returns tracks ordered by play count, then newest first to
// break ties. A positive days restricts to tracks created within that window;
// days <= 0 spans all time.
func (r *trackRepo) GetTrendingTracks(ctx context.Context, limit int, offset int, days int) ([]*Track, error) {
	var tracks []*Track

	query := r.Db.WithContext(ctx).
		Preload("Artist", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "created_at", "updated_at")
		})

	if days > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -days)
		query = query.Where("created_at >= ?", cutoffDate)
	}

	res := query.
		Order("play_count DESC").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tracks)

	if res.Error != nil {
		return nil, res.Error
	}

	return tracks, nil
}

// GetRecentTracks returns tracks newest first by creation time.
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

// IncrementPlayCount atomically bumps the track's play count in the database
// (no read-modify-write). A missing track id is not an error: zero rows update.
func (r *trackRepo) IncrementPlayCount(ctx context.Context, trackId uuid.UUID) error {
	res := r.Db.WithContext(ctx).
		Model(&Track{}).
		Where("id = ?", trackId).
		UpdateColumn("play_count", gorm.Expr("play_count + ?", 1))

	return res.Error
}

// RecordPlayback appends a playback-history row stamped with the current time;
// durationPlayed is the seconds listened. It does not touch the track's play count.
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
