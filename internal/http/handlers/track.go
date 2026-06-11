package handlers

import (
	"auxstream/internal/cache"
	"auxstream/internal/db"
	fs "auxstream/internal/storage"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// GetTrackByIDHandler fetches a single track by ID
func GetTrackByIDHandler(c *gin.Context, r db.TrackRepo) {
	trackIdStr := c.Param("id")
	trackId, err := uuid.Parse(trackIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid track ID format"))
		return
	}

	track, err := r.GetTrackByID(c, trackId)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("track not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": track,
	})
}

// FetchTracksByArtistHandler fetch tracks by artist (limit results < 100)
func FetchTracksByArtistHandler(c *gin.Context, r db.TrackRepo) {
	artist := c.Query("artist")
	tracks, err := r.GetTrackByArtist(c, artist)
	if err != nil {
		log.Printf("GetTrackByArtist error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to fetch tracks"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
	})

}

type FetchTrackQueryParams struct {
	PageSize int    `form:"pagesize" binding:"gte=0"`
	PageNum  int    `form:"pagenumber" binding:"gte=1"`
	Sort     string `form:"sort"` // "trending", "recent", or default
	Days     int    `form:"days"` // For trending within last N days (0 = all time)
}

// FetchTracksHandler fetch paginated tracks with limit on page size and sorting options
func FetchTracksHandler(c *gin.Context, r db.TrackRepo) {
	var reqParams FetchTrackQueryParams

	if err := c.ShouldBindQuery(&reqParams); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	limit := reqParams.PageSize
	offset := (reqParams.PageNum - 1) * reqParams.PageSize

	var tracks []*db.Track
	var err error

	// Choose sorting method based on query parameter
	switch reqParams.Sort {
	case "trending":
		days := reqParams.Days
		if days == 0 {
			days = 30 // Default to last 30 days for trending
		}
		tracks, err = r.GetTrendingTracks(c, limit, offset, days)
	case "recent":
		tracks, err = r.GetRecentTracks(c, limit, offset)
	default:
		// Default behavior - just get tracks (no special sorting)
		tracks, err = r.GetTracks(c, limit, offset)
	}

	if err != nil {
		log.Printf("FetchTracks error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse("failed to fetch tracks"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
	})

}

type AddTrackForm struct {
	Title     string                `form:"title" binding:"required"`
	ArtistId  string                `form:"artist_id" binding:"required"`
	Audio     *multipart.FileHeader `form:"audio" binding:"required"`
	Duration  int                   `form:"duration"`  // Optional: duration in seconds
	Thumbnail string                `form:"thumbnail"` // Optional: thumbnail URL or path
}

// AddTrackHandler add track to the system
func AddTrackHandler(c *gin.Context, r db.TrackRepo, artistRepo db.ArtistRepo) {
	var reqForm AddTrackForm
	if err := c.ShouldBind(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	trackTitle := reqForm.Title

	trackArtistID, err := uuid.Parse(reqForm.ArtistId)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Sprintf("artist id should be a valid uuid string not %s", reqForm.ArtistId)))
		return
	}

	file := reqForm.Audio

	if file.Size <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("audio for track not found"))
		return
	}

	if !strings.HasSuffix(file.Filename, ".mp3") {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("invalid audio format use mp3 instead"))
		return
	}

	audioFile, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("unable to access track audio"))
		return
	}
	defer audioFile.Close()
	// io.ReadFull guarantees the whole file is read; a bare Read may return fewer
	// bytes than requested and silently truncate the stored audio.
	audioBytes := make([]byte, file.Size)
	if _, readErr := io.ReadFull(audioFile, audioBytes); readErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("unable to read track audio"))
		return
	}
	filePath, err := fs.Store.Save(audioBytes)
	if err != nil {
		log.Printf("store audio error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse("failed to store audio"))
		return
	}

	artist := &db.Artist{ID: trackArtistID}

	ctx := c.Request.Context()
	cacheClient, ok := ctx.Value(CacheContextKey).(cache.Cache)

	artistCacheKey := fmt.Sprintf("artist-id-%s", trackArtistID)
	if ok {
		if cacheErr := cacheClient.Get(artistCacheKey, artist); cacheErr != nil {
			log.Printf("(Get artist id from cache) failed: %s\n", cacheErr.Error())
		}
	}

	if artist.Name == "" {
		artist, err = artistRepo.GetArtistById(c, trackArtistID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errorResponse(fmt.Sprintf("artist with id (%s) does not exist: %s", trackArtistID, err.Error())))
			return
		}
		if ok {
			_ = cacheClient.Set(artistCacheKey, artist, 10*time.Hour)
		}
	}

	track, err := r.CreateTrack(c, trackTitle, trackArtistID, filePath, reqForm.Duration, reqForm.Thumbnail)
	if err != nil {
		log.Printf("create track error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse("failed to save track"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": track,
	})
}

type BulkTrackUploadForm struct {
	Titles   []string                `form:"track_titles" binding:"required"`
	Files    []*multipart.FileHeader `form:"track_files" binding:"required"`
	ArtistId string                  `form:"artist_id" binding:"required"`
}

// BulkTrackUploadHandler enables bulk track uploads
func BulkTrackUploadHandler(c *gin.Context, r db.TrackRepo) {
	var reqForm BulkTrackUploadForm

	if err := c.ShouldBind(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	artistID, err := uuid.Parse(reqForm.ArtistId)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Sprintf("invalid artist id: %s", reqForm.ArtistId)))
		return
	}

	fileMetas, err := processFiles(reqForm.Files, reqForm.Titles)
	if err != nil {
		log.Printf("bulk process files error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse("audio upload failed"))
		return
	}

	var trackTitles []string
	filteredFileNames := map[string]string{}
	for title, fileMeta := range fileMetas {
		if fileMeta.Name != "" {
			trackTitles = append(trackTitles, title)
			filteredFileNames[title] = fileMeta.Name
		}
	}
	rows, err := r.BulkCreateTracks(c, trackTitles, artistID, filteredFileNames)

	if err != nil {
		log.Printf("bulk create tracks error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse("audio upload failed"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": map[string]any{
			"saved": filteredFileNames,
			"rows":  rows,
		},
	})
}

func processFiles(files []*multipart.FileHeader, titles []string) (map[string]fs.FileMeta, error) {
	var groupFiles []fs.FileMeta
	groupTitles := map[string]string{}

	for idx, file := range files {
		groupTitles[file.Filename] = titles[idx]

		audioFile, fileErr := file.Open()
		if fileErr != nil {
			groupFiles = append(groupFiles, fs.FileMeta{AudioTitle: file.Filename, Content: []byte{}})
			continue
		}
		audioBytes := make([]byte, file.Size)
		_, readErr := io.ReadFull(audioFile, audioBytes)
		_ = audioFile.Close()
		if readErr != nil {
			groupFiles = append(groupFiles, fs.FileMeta{AudioTitle: file.Filename, Content: []byte{}})
			continue
		}
		groupFiles = append(groupFiles, fs.FileMeta{AudioTitle: file.Filename, Content: audioBytes})
	}

	resultCh := make(chan fs.FileMeta, len(groupFiles))
	fs.Store.BulkSave(resultCh, groupFiles)

	fileMeta := map[string]fs.FileMeta{}
	for fmeta := range resultCh {
		if fmeta.Name != "" {
			fileMeta[groupTitles[fmeta.AudioTitle]] = fmeta
		}
	}

	return fileMeta, nil
}

type TrackPlayRequest struct {
	TrackID        string `json:"track_id" binding:"required"`
	DurationPlayed int    `json:"duration_played"` // Optional: how long the user listened (seconds)
}

// TrackPlayHandler records a track play/listen event
func TrackPlayHandler(c *gin.Context, r db.TrackRepo) {
	var req TrackPlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	trackId, err := uuid.Parse(req.TrackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid track ID"))
		return
	}

	if err := r.IncrementPlayCount(c, trackId); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("failed to record play"))
		return
	}

	// Optionally record in playback history if user is authenticated
	userIdValue, exists := c.Get("user_id")
	if exists {
		if userId, ok := userIdValue.(uuid.UUID); ok {
			duration := req.DurationPlayed
			if duration == 0 {
				duration = 30 // Default minimum listening time
			}
			_ = r.RecordPlayback(c, userId, trackId, duration)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "play recorded",
	})
}
