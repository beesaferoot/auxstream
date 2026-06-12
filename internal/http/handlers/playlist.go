package handlers

import (
	"auxstream/internal/auth"
	"auxstream/internal/db"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// playlistResponse is the trimmed playlist shape returned to clients.
type playlistResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	TrackCount  int64     `json:"track_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// playlistDetailResponse adds the ordered track list for the detail view.
type playlistDetailResponse struct {
	playlistResponse
	Tracks []*db.Track `json:"tracks"`
}

func toPlaylistResponse(p *db.Playlist, trackCount int64) playlistResponse {
	return playlistResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsPublic:    p.IsPublic,
		TrackCount:  trackCount,
		CreatedAt:   p.CreatedAt,
	}
}

// currentUserID pulls the authenticated user's id from context (set by the JWT middleware).
func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	claims, ok := auth.GetUserFromContext(c)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func parsePlaylistID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid playlist id"))
		return uuid.Nil, false
	}
	return id, true
}

// ownedPlaylistOr404 loads the playlist and confirms the caller owns it. On any
// miss it writes a 404 (never leaking whether someone else's playlist exists) and
// returns ok=false.
func ownedPlaylistOr404(c *gin.Context, r db.PlaylistRepo) (*db.Playlist, bool) {
	id, ok := parsePlaylistID(c)
	if !ok {
		return nil, false
	}
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("authentication required"))
		return nil, false
	}
	p, err := r.GetPlaylistByID(c.Request.Context(), id)
	if err != nil || p.UserID != userID {
		c.JSON(http.StatusNotFound, errorResponse("playlist not found"))
		return nil, false
	}
	return p, true
}

// GetUserPlaylistsHandler returns the authenticated user's playlists with track counts.
func GetUserPlaylistsHandler(c *gin.Context, r db.PlaylistRepo) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("authentication required"))
		return
	}

	playlists, err := r.GetUserPlaylists(c.Request.Context(), userID)
	if err != nil {
		log.Printf("GetUserPlaylists error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to fetch playlists"))
		return
	}

	resp := make([]playlistResponse, len(playlists))
	for i, p := range playlists {
		resp[i] = toPlaylistResponse(&p.Playlist, p.TrackCount)
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

type createPlaylistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// CreatePlaylistHandler creates a playlist owned by the authenticated caller from
// a JSON body (name required); responds 201.
func CreatePlaylistHandler(c *gin.Context, r db.PlaylistRepo) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("authentication required"))
		return
	}
	var req createPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	p, err := r.CreatePlaylist(c.Request.Context(), userID, req.Name, req.Description, req.IsPublic)
	if err != nil {
		log.Printf("CreatePlaylist error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to create playlist"))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": toPlaylistResponse(p, 0)})
}

// GetPlaylistHandler returns a playlist with its tracks. Readable by the owner, or by
// anyone when the playlist is public (shared link). Uses optional auth.
func GetPlaylistHandler(c *gin.Context, r db.PlaylistRepo) {
	id, ok := parsePlaylistID(c)
	if !ok {
		return
	}
	p, err := r.GetPlaylistByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("playlist not found"))
		return
	}
	if !p.IsPublic {
		userID, authed := currentUserID(c)
		if !authed || p.UserID != userID {
			c.JSON(http.StatusNotFound, errorResponse("playlist not found"))
			return
		}
	}

	tracks, err := r.GetPlaylistTracks(c.Request.Context(), id)
	if err != nil {
		log.Printf("GetPlaylistTracks error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to fetch playlist"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": playlistDetailResponse{
		playlistResponse: toPlaylistResponse(p, int64(len(tracks))),
		Tracks:           tracks,
	}})
}

type updatePlaylistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// UpdatePlaylistHandler replaces a playlist's fields from a JSON body. Restricted
// to the owner; non-owners and unknown ids alike get 404.
func UpdatePlaylistHandler(c *gin.Context, r db.PlaylistRepo) {
	p, ok := ownedPlaylistOr404(c, r)
	if !ok {
		return
	}
	var req updatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	updated, err := r.UpdatePlaylist(c.Request.Context(), p.ID, req.Name, req.Description, req.IsPublic)
	if err != nil {
		log.Printf("UpdatePlaylist error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to update playlist"))
		return
	}
	tracks, _ := r.GetPlaylistTracks(c.Request.Context(), p.ID)
	c.JSON(http.StatusOK, gin.H{"data": toPlaylistResponse(updated, int64(len(tracks)))})
}

// DeletePlaylistHandler deletes an owned playlist; non-owners and unknown ids get 404.
func DeletePlaylistHandler(c *gin.Context, r db.PlaylistRepo) {
	p, ok := ownedPlaylistOr404(c, r)
	if !ok {
		return
	}
	if err := r.DeletePlaylist(c.Request.Context(), p.ID); err != nil {
		log.Printf("DeletePlaylist error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to delete playlist"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "playlist deleted"})
}

type addTrackRequest struct {
	TrackID string `json:"track_id" binding:"required"`
}

// AddTrackToPlaylistHandler appends a track (track_id in the JSON body) to an
// owned playlist. The track must exist (else 404), as must the playlist.
func AddTrackToPlaylistHandler(c *gin.Context, r db.PlaylistRepo, trackRepo db.TrackRepo) {
	p, ok := ownedPlaylistOr404(c, r)
	if !ok {
		return
	}
	var req addTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	trackID, err := uuid.Parse(req.TrackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid track id"))
		return
	}
	if _, err := trackRepo.GetTrackByID(c.Request.Context(), trackID); err != nil {
		c.JSON(http.StatusNotFound, errorResponse("track not found"))
		return
	}
	if err := r.AddTrack(c.Request.Context(), p.ID, trackID); err != nil {
		log.Printf("AddTrack error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to add track"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "track added"})
}

// RemoveTrackHandler drops the track named by the trackId path param from an
// owned playlist.
func RemoveTrackHandler(c *gin.Context, r db.PlaylistRepo) {
	p, ok := ownedPlaylistOr404(c, r)
	if !ok {
		return
	}
	trackID, err := uuid.Parse(c.Param("trackId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid track id"))
		return
	}
	if err := r.RemoveTrack(c.Request.Context(), p.ID, trackID); err != nil {
		log.Printf("RemoveTrack error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to remove track"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "track removed"})
}

type reorderTracksRequest struct {
	TrackIDs []string `json:"track_ids" binding:"required"`
}

// ReorderTracksHandler sets a playlist's track order from the track_ids body
// array, which is expected to be the full ordering. Owner only; any unparseable
// id rejects the whole request with 400.
func ReorderTracksHandler(c *gin.Context, r db.PlaylistRepo) {
	p, ok := ownedPlaylistOr404(c, r)
	if !ok {
		return
	}
	var req reorderTracksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	ordered := make([]uuid.UUID, 0, len(req.TrackIDs))
	for _, s := range req.TrackIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("invalid track id in order"))
			return
		}
		ordered = append(ordered, id)
	}
	if err := r.ReorderTracks(c.Request.Context(), p.ID, ordered); err != nil {
		log.Printf("ReorderTracks error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to reorder tracks"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tracks reordered"})
}
