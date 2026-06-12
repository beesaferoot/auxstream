package db

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlaylistWithCount is a playlist plus its (non-deleted) track count, for list views.
type PlaylistWithCount struct {
	Playlist
	TrackCount int64
}

type PlaylistRepo interface {
	GetUserPlaylists(ctx context.Context, userID uuid.UUID) ([]PlaylistWithCount, error)
	GetPlaylistByID(ctx context.Context, id uuid.UUID) (*Playlist, error)
	GetPlaylistTracks(ctx context.Context, playlistID uuid.UUID) ([]*Track, error)
	CreatePlaylist(ctx context.Context, userID uuid.UUID, name, description string, isPublic bool) (*Playlist, error)
	UpdatePlaylist(ctx context.Context, id uuid.UUID, name, description string, isPublic bool) (*Playlist, error)
	DeletePlaylist(ctx context.Context, id uuid.UUID) error
	AddTrack(ctx context.Context, playlistID, trackID uuid.UUID) error
	RemoveTrack(ctx context.Context, playlistID, trackID uuid.UUID) error
	ReorderTracks(ctx context.Context, playlistID uuid.UUID, orderedTrackIDs []uuid.UUID) error
}

type playlistRepo struct {
	Db *gorm.DB
}

func NewPlaylistRepo(db *gorm.DB) PlaylistRepo {
	return &playlistRepo{Db: db}
}

// GetUserPlaylists returns the user's playlists (newest first) with their track counts.
func (r *playlistRepo) GetUserPlaylists(ctx context.Context, userID uuid.UUID) ([]PlaylistWithCount, error) {
	var playlists []Playlist
	if err := r.Db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&playlists).Error; err != nil {
		return nil, err
	}
	if len(playlists) == 0 {
		return []PlaylistWithCount{}, nil
	}

	ids := make([]uuid.UUID, len(playlists))
	for i, p := range playlists {
		ids[i] = p.ID
	}

	// One grouped count query for all the user's playlists (avoids N+1).
	type countRow struct {
		PlaylistID uuid.UUID
		Count      int64
	}
	var rows []countRow
	if err := r.Db.WithContext(ctx).
		Model(&PlaylistTrack{}).
		Select("playlist_id, count(*) as count").
		Where("playlist_id IN ?", ids).
		Group("playlist_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	countByID := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		countByID[row.PlaylistID] = row.Count
	}

	out := make([]PlaylistWithCount, len(playlists))
	for i, p := range playlists {
		out[i] = PlaylistWithCount{Playlist: p, TrackCount: countByID[p.ID]}
	}
	return out, nil
}

func (r *playlistRepo) GetPlaylistByID(ctx context.Context, id uuid.UUID) (*Playlist, error) {
	var p Playlist
	res := r.Db.WithContext(ctx).First(&p, "id = ?", id)
	return &p, res.Error
}

// GetPlaylistTracks returns the playlist's tracks in playlist order (position, then added).
func (r *playlistRepo) GetPlaylistTracks(ctx context.Context, playlistID uuid.UUID) ([]*Track, error) {
	var entries []PlaylistTrack
	if err := r.Db.WithContext(ctx).
		Where("playlist_id = ?", playlistID).
		Order("position ASC, added_at ASC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return []*Track{}, nil
	}

	ids := make([]uuid.UUID, len(entries))
	for i, e := range entries {
		ids[i] = e.TrackID
	}

	var tracks []*Track
	if err := r.Db.WithContext(ctx).
		Preload("Artist").
		Where("id IN ?", ids).
		Find(&tracks).Error; err != nil {
		return nil, err
	}

	// Restore playlist order (the IN query doesn't preserve it).
	byID := make(map[uuid.UUID]*Track, len(tracks))
	for _, t := range tracks {
		byID[t.ID] = t
	}
	ordered := make([]*Track, 0, len(entries))
	for _, e := range entries {
		if t, ok := byID[e.TrackID]; ok {
			ordered = append(ordered, t)
		}
	}
	return ordered, nil
}

func (r *playlistRepo) CreatePlaylist(ctx context.Context, userID uuid.UUID, name, description string, isPublic bool) (*Playlist, error) {
	p := &Playlist{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
	}
	res := r.Db.WithContext(ctx).Create(p)
	return p, res.Error
}

func (r *playlistRepo) UpdatePlaylist(ctx context.Context, id uuid.UUID, name, description string, isPublic bool) (*Playlist, error) {
	// Map (not struct) so is_public=false is written rather than skipped as a zero value.
	if err := r.Db.WithContext(ctx).
		Model(&Playlist{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"name":        name,
			"description": description,
			"is_public":   isPublic,
		}).Error; err != nil {
		return nil, err
	}
	return r.GetPlaylistByID(ctx, id)
}

// DeletePlaylist soft-deletes the playlist and its track entries in one transaction.
func (r *playlistRepo) DeletePlaylist(ctx context.Context, id uuid.UUID) error {
	return r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("playlist_id = ?", id).Delete(&PlaylistTrack{}).Error; err != nil {
			return err
		}
		return tx.Delete(&Playlist{}, "id = ?", id).Error
	})
}

// AddTrack appends a track to the end of the playlist. Adding a track that's already
// present is a no-op (guarded by the partial unique index).
func (r *playlistRepo) AddTrack(ctx context.Context, playlistID, trackID uuid.UUID) error {
	var existing int64
	if err := r.Db.WithContext(ctx).
		Model(&PlaylistTrack{}).
		Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	var maxPos struct{ Max int }
	if err := r.Db.WithContext(ctx).
		Model(&PlaylistTrack{}).
		Select("COALESCE(MAX(position), 0) as max").
		Where("playlist_id = ?", playlistID).
		Scan(&maxPos).Error; err != nil {
		return err
	}

	entry := &PlaylistTrack{
		ID:         uuid.New(),
		PlaylistID: playlistID,
		TrackID:    trackID,
		Position:   maxPos.Max + 1,
		AddedAt:    time.Now(),
	}
	if err := r.Db.WithContext(ctx).Create(entry).Error; err != nil {
		// Lost a race against a concurrent add — the unique index rejected it; treat as a no-op.
		if isUniqueViolation(err) {
			return nil
		}
		return err
	}
	return nil
}

func (r *playlistRepo) RemoveTrack(ctx context.Context, playlistID, trackID uuid.UUID) error {
	return r.Db.WithContext(ctx).
		Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
		Delete(&PlaylistTrack{}).Error
}

// ReorderTracks assigns positions 1..n matching the given track order.
func (r *playlistRepo) ReorderTracks(ctx context.Context, playlistID uuid.UUID, orderedTrackIDs []uuid.UUID) error {
	return r.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, trackID := range orderedTrackIDs {
			if err := tx.Model(&PlaylistTrack{}).
				Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
				Update("position", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "23505")
}
